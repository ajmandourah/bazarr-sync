/*
Copyright © 2024 ajmandourah
*/
package cli

import (
	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"os"
	"fmt"
	"strconv"

	"github.com/pterm/pterm"
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

	showsCmd.Flags().IntSliceVar(&sonarrid,"sonarr-id",[]int{},"Specify a list of sonarr Ids to sync. Use --list to view your shows with respective sonarr id.")
	showsCmd.Flags().IntVar(&showsContinueFrom,"continue-from",-1,"Continue with the given Sonarr episode ID.")
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
	for i, show := range shows.Data {

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
				label := pterm.LightBlue(show.Title+":"+episode.Title)+ " lang:"+pterm.LightRed(subtitle.Code2)+" "+strconv.Itoa(i+1)+"/"+strconv.Itoa(len(shows.Data))

				if isTerminal {
					p,_ := pterm.DefaultSpinner.Start(label)

					if skipForward {
						if episode.SonarrEpisodeId == showsContinueFrom {
							skipForward = false
						} else {
							pterm.Success.Prefix = pterm.Prefix{Text: "SKIP", Style: pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack)}
							p.Success("Skipping due to continue option.")
							pterm.Success.Prefix = pterm.Prefix{Text: "SUCCESS", Style: pterm.NewStyle(pterm.BgGreen, pterm.FgBlack)}
							stats.skipped++
							continue
						}
					}

					c <- episode.SonarrEpisodeId
					if subtitle.Path == "" || subtitle.File_size == 0 {
						pterm.Success.Prefix = pterm.Prefix{Text: "SKIP", Style: pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack)}
						p.Success("Could not find a subtitle. most likely it is embedded. Lang: ",subtitle.Code2)
						pterm.Success.Prefix = pterm.Prefix{Text: "SUCCESS", Style: pterm.NewStyle(pterm.BgGreen, pterm.FgBlack)}
						stats.skipped++
						continue
					}
					params := bazarr.GetSyncParams("episode", episode.SonarrEpisodeId, subtitle)
					if gss {params.Gss = "True"}
					if no_framerate_fix {params.No_framerate_fix = "True"}
					ok := bazarr.Sync(cfg, params)
					if ok {
						p.Success("Synced lang: "+subtitle.Code2)
						stats.success++
						continue
					} else {
						// Retry with exponential backoff
						p.Warning("Error while syncing lang: "+subtitle.Code2+" Retrying..")
						ok = retrySync(cfg, params, show.Title+": "+episode.Title, subtitle.Code2)
						if ok {
							p.Success("Synced lang: "+subtitle.Code2)
							stats.success++
						} else {
							p.Fail("Unable to sync lang: "+subtitle.Code2)
							stats.failed++
						}
					}
				} else {
					// Non-TTY: simple text output, no spinner animation
					if skipForward {
						if episode.SonarrEpisodeId == showsContinueFrom {
							skipForward = false
						} else {
							fmt.Printf("  SKIP %s:%s (continue)\n", show.Title, episode.Title)
							stats.skipped++
							continue
						}
					}

					c <- episode.SonarrEpisodeId
					if subtitle.Path == "" || subtitle.File_size == 0 {
						fmt.Printf("  SKIP %s:%s lang=%s (no subtitle file)\n", show.Title, episode.Title, subtitle.Code2)
						stats.skipped++
						continue
					}
					params := bazarr.GetSyncParams("episode", episode.SonarrEpisodeId, subtitle)
					if gss {params.Gss = "True"}
					if no_framerate_fix {params.No_framerate_fix = "True"}
					ok := bazarr.Sync(cfg, params)
					if ok {
						fmt.Printf("  SYNCED %s:%s lang=%s\n", show.Title, episode.Title, subtitle.Code2)
						stats.success++
						continue
					} else {
						fmt.Printf("  FAILED %s:%s lang=%s\n", show.Title, episode.Title, subtitle.Code2)
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
		fmt.Fprintln(os.Stderr, "Query Error: Could not query shows:", err)
		os.Exit(1)
	}
	table := pterm.TableData{
		{"Title","SonarrSeriesId"},
	}
	pterm.Println(pterm.LightGreen("Listing all your Series with their respective Sonarr ID (great for syncing specific series)\n"))

	for _, show := range shows.Data {
		// pterm.Println(pterm.LightBlue(movie.Title), "\t", pterm.LightRed(movie.ImdbId))
		t := []string{pterm.LightBlue(show.Title),pterm.LightRed(show.SonarrSeriesId)}
		table = append(table, t)
	}
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(table).Render()
}
