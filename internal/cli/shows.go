/*
Copyright © 2024 ajmandourah
*/
package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	"github.com/briandowns/spinner"
	"github.com/pterm/pterm"
	"golang.org/x/term"

	"github.com/spf13/cobra"
)

var sonarrid []int
var showsContinueFrom int

// showsCmd represents the shows command
var showsCmd = &cobra.Command{
	Use:   "shows",
	Short: "Sync subtitles to the audio track of the show's episodes",
	Example: "bazarr-sync --config config.yaml sync shows --no-framerate-fix",
	Long: `By default, Bazarr will try to sync the sub to the audio track:0 of the media. 
This can fail due to many reasons mainly due to failure of bazarr to extract audio info. This is unfortunatly out of my hands.
The script by default will try to not use the golden section search method and will try to fix framerate issues. This can be changed using the flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()
		bazarr.HealthCheck(cfg)
		if to_list {
			list_shows(cfg)
			return
		}
		runWithSignalHandler(func(c chan int){
			sync_shows(cfg, c)
		})
	},
}

func init() {
	syncCmd.AddCommand(showsCmd)

	showsCmd.Flags().IntSliceVar(&sonarrid,"sonarr-id",[]int{},`Specify a list of sonarr Ids to sync. Use --list to view your shows with respective sonarr id.`)
	showsCmd.Flags().IntVar(&showsContinueFrom,"continue-from",-1,"Continue with the given Sonarr episode ID.")
}

// startShowSpinner creates and starts a spinner with plain label suffix (no ANSI codes — they break \r cursor tracking).
func startShowSpinner(label string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
	s.Writer = os.Stderr
	s.Suffix = " " + label // plain text, no colors — safe for \r redraws
	s.Start()
	return s
}

// stopShowSpinner sets FinalMSG with colored result and stops the spinner.
func stopShowSpinner(s *spinner.Spinner, label string, green bool) {
	if green {
		s.FinalMSG = fmt.Sprintf("  %s%s\n", pterm.LightGreen("✅ "), pterm.LightGreen(label))
	} else {
		s.FinalMSG = fmt.Sprintf("  %s%s\n", pterm.LightRed("❌ "), pterm.LightRed(label))
	}
	s.Stop()
}

func sync_shows(cfg config.Config, c chan int) {
	shows, err := bazarr.QuerySeries(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query series:", err)
		os.Exit(1)
	}
	fmt.Println("Syncing shows in your Bazarr library.")

	skipForward := showsContinueFrom != -1
	stats := syncStats{}
	isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
showsLoop:
	for _, show := range shows.Data {

		if len(sonarrid) > 0 {
			for _, id := range sonarrid {
				if id == show.SonarrSeriesId {
					goto episodes
				}
			}
			continue showsLoop
		}
	
episodes:
		episodes, err := bazarr.QueryEpisodes(cfg,show.SonarrSeriesId)
		if err != nil {
			continue
		}

		for _, episode := range episodes.Data {
			for _, subtitle := range episode.Subtitles {
				// Language filtering
				if lang != "" && subtitle.Code2 != lang {
					continue
				}

				label := fmt.Sprintf("lang:%s S%02dE%02d %s", subtitle.Code2, episode.SeasonNumber, episode.EpisodeNumber, show.Title)

				if isTerminal {
					s := startShowSpinner(label)

					if skipForward {
						if episode.SonarrEpisodeId == showsContinueFrom {
							skipForward = false
						} else {
							stopShowSpinner(s, label, false) // ❌ red — skipped
							stats.skipped++
							continue
						}
					}

					c <- episode.SonarrEpisodeId
					if subtitle.Path == "" || subtitle.File_size == 0 {
						stopShowSpinner(s, label, false) // ❌ red — no sub file
						stats.skipped++
						continue
					}
					params := bazarr.GetSyncParams("episode", episode.SonarrEpisodeId, subtitle)
					if gss {params.Gss = "True"}
					if no_framerate_fix {params.No_framerate_fix = "True"}
					ok := bazarr.Sync(cfg, params)
					if ok {
						stopShowSpinner(s, label, true) // ✅ green — success
						stats.success++
						continue
					} else {
						// Retry with exponential backoff
						fmt.Fprint(os.Stderr, "\r\033[K") // clear spinner line for warning
						fmt.Fprintf(os.Stderr, "  WARNING: Error while syncing lang:%s\n", subtitle.Code2)
						ok = retrySync(cfg, params, show.Title+": "+episode.Title, subtitle.Code2)
						if ok {
							stopShowSpinner(s, label, true) // ✅ green — retry success
							stats.success++
						} else {
							stopShowSpinner(s, label, false) // ❌ red — hard failure
							stats.failed++
						}
					}
				} else {
					// Non-TTY: simple text output, no spinner animation
					fmt.Printf("  %s\n", label)

					if skipForward {
						if episode.SonarrEpisodeId == showsContinueFrom {
							skipForward = false
						} else {
							stats.skipped++
							continue
						}
					}

					c <- episode.SonarrEpisodeId
					if subtitle.Path == "" || subtitle.File_size == 0 {
						pterm.Info.Printf("    (no sub file - probably embedded)\n")
						stats.skipped++
						continue
					}
					params := bazarr.GetSyncParams("episode", episode.SonarrEpisodeId, subtitle)
					if gss {params.Gss = "True"}
					if no_framerate_fix {params.No_framerate_fix = "True"}
					ok := bazarr.Sync(cfg, params)
					if ok {
						fmt.Printf("  %s\n", pterm.LightGreen("[Request sent]"))
						stats.success++
						continue
					} else {
						fmt.Printf("  %s\n", pterm.LightRed("[Error]"))
						stats.failed++
					}
				}
			}
		}
	}
	fmt.Println("Finished syncing subtitles of type Shows")
	fmt.Printf("\n📊 Summary: %d synced, %d skipped, %d failed\n", stats.success, stats.skipped, stats.failed)
	// Signal that we're done with all subtitles.
	close(c)
}


func list_shows(cfg config.Config) {
	shows, err := bazarr.QuerySeries(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query series:", err)
		os.Exit(1)
	}
	table := pterm.TableData{
		{"Title","SonarrSeriesId"},
	}
	pterm.Println(pterm.LightGreen("Listing all your Series with their respective Sonarr ID (great for syncing specific series)\n"))

	for _, show := range shows.Data {
		t := []string{pterm.LightBlue(show.Title),pterm.LightRed(show.SonarrSeriesId)}
		table = append(table, t)
	}
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(table).Render()
}
