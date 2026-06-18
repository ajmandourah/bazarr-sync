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

var showsCmd = &cobra.Command{
	Use:   "shows",
	Short: "Sync subtitles to the audio track of the show's episodes",
	Example: "bazarr-sync --config config.yaml sync shows --no-framerate-fix",
	Long: `By default, Bazarr will try to sync the sub to the audio track:0 of the media. 
This can fail due to many reasons mainly due to failure of bazarr to extract audio info. This is unfortunatly out of my hands.
The script by default will try to not use the golden section search method and will try to fix framerate issues. This can be changed using the flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()
		if _, err := bazarr.CheckHealth(cfg); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
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

func startShowSpinner(label string) *spinner.Spinner {
	s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
	s.Writer = os.Stderr
	s.Suffix = " " + label
	s.Start()
	return s
}

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
				if id == show.SonarrId {
					goto episodes
				}
			}
			continue showsLoop
		}
	
episodes:
		episodes, err := bazarr.QueryEpisodes(cfg, show.SonarrId)
		if err != nil {
			continue
		}

		for _, episode := range episodes.Data {
			for _, subtitle := range episode.Subtitles {
				if lang != "" && subtitle.Code2 != lang {
					continue
				}

				label := fmt.Sprintf("lang:%s S%02dE%02d %s", subtitle.Code2, episode.SeasonNumber, episode.EpisodeNumber, show.Title)

				if isTerminal {
					s := startShowSpinner(label)

					if skipForward {
						if episode.SonarrEpId == showsContinueFrom {
							skipForward = false
						} else {
							stopShowSpinner(s, label, false)
							stats.skipped++
							continue
						}
					}

					c <- episode.SonarrEpId
					if subtitle.Path == "" || subtitle.FileSize == 0 {
						stopShowSpinner(s, label, false)
						stats.skipped++
						continue
					}
					params := bazarr.GetSyncParams("episode", episode.SonarrEpId, subtitle)
					if gss { params.Gss = "True" }
					if no_framerate_fix { params.NoFramerateFix = "True" }
					ok := bazarr.Sync(cfg, params)
					if ok {
						stopShowSpinner(s, label, true)
						stats.success++
						continue
					} else {
						fmt.Fprint(os.Stderr, "\r\033[K")
						fmt.Fprintf(os.Stderr, "  WARNING: Error while syncing lang:%s\n", subtitle.Code2)
						ok = retrySync(cfg, params, show.Title+": "+episode.Title, subtitle.Code2)
						if ok {
							stopShowSpinner(s, label, true)
							stats.success++
						} else {
							stopShowSpinner(s, label, false)
							stats.failed++
						}
					}
				} else {
					fmt.Printf("  %s\n", label)

					if skipForward {
						if episode.SonarrEpId == showsContinueFrom {
							skipForward = false
						} else {
							stats.skipped++
							continue
						}
					}

					c <- episode.SonarrEpId
					if subtitle.Path == "" || subtitle.FileSize == 0 {
						pterm.Info.Printf("    (no sub file - probably embedded)\n")
						stats.skipped++
						continue
					}
					params := bazarr.GetSyncParams("episode", episode.SonarrEpId, subtitle)
					if gss { params.Gss = "True" }
					if no_framerate_fix { params.NoFramerateFix = "True" }
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
	close(c)
}

func list_shows(cfg config.Config) {
	shows, err := bazarr.QuerySeries(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query series:", err)
		os.Exit(1)
	}
	table := pterm.TableData{
		{"Title","SonarrId"},
	}
	pterm.Println(pterm.LightGreen("Listing all your Series with their respective Sonarr ID (great for syncing specific series)\n"))

	for _, show := range shows.Data {
		t := []string{pterm.LightBlue(show.Title), pterm.LightRed(show.SonarrId)}
		table = append(table, t)
	}
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(table).Render()
}
