/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"ajmandourah/bazarr-sync/internal/bazarr"
	"ajmandourah/bazarr-sync/internal/config"
	"github.com/spf13/cobra"
	"os"
	"fmt"
	"strconv"
	"time"


	"github.com/schollz/progressbar/v3"
)

// showsCmd represents the shows command
var showsCmd = &cobra.Command{
	Use:   "shows",
	Short: "Sync subtitles to the audio track of the show's episodes",
	Example: "bazarr-sync --config config.yaml sync movies --no-framerate-fix",
	Long: `By default, Bazarr will try to sync the sub to the audio track:0 of the media. 
This can fail due to many reasons mainly due to failure of bazarr to extract audio info. This is unfortunatly out of my hands.
The script by default will try to not use the golden section search method and will try to fix framerate issues. This can be changed using the flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()
		bazarr.HealthCheck(cfg)
		sync_shows(cfg)
	},
}

func init() {
	syncCmd.AddCommand(showsCmd)
}

func sync_shows(cfg config.Config) {
	shows, err := bazarr.QuerySeries(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query series")
	}
	fmt.Println("Syncing ", len(shows.Data), "shows in your Bazarr library.")

	bar := progressbar.NewOptions(len(shows.Data),
					progressbar.OptionFullWidth(),
					progressbar.OptionClearOnFinish(),
					progressbar.OptionShowCount())
	for _, show := range shows.Data {
		bar.Add(1)
		bar.Describe(show.Title)
		episodes, err := bazarr.QueryEpisodes(cfg,show.SonarrSeriesId)
		if err != nil {
			bar.Clear()
			fmt.Println("Could not Query episodes for show:", show.Title)
			bar.RenderBlank()
			continue
		}

		for _, episode := range episodes.Data {
			for sub_index, subtitle := range episode.Subtitles {
				if subtitle.Path == "" || subtitle.File_size == 0 {
					continue
				}
				bar.Describe(show.Title + ":" + episode.Title + " " + strconv.Itoa(sub_index+1) + "/" + strconv.Itoa(len(episode.Subtitles)))
				params := bazarr.GetSyncParams("episode", episode.SonarrEpisodeId, subtitle)
				if gss {params.Gss = "True"}
				if no_framerate_fix {params.No_framerate_fix = "True"}
				ok := bazarr.Sync(cfg, params)
				if !ok {
					for i := 1; i < 5; i++ {
						time.Sleep(2 * time.Second)
						ok := bazarr.Sync(cfg, params)
						if ok {
							break
						}
					}
					bar.Clear()
					fmt.Println("Unable to sync subtitile for", show.Title,":", episode.Title, "lang:", subtitle.Code2)
					bar.RenderBlank()
				}
			}
		}
	}
	bar.Clear()
	fmt.Println("Finished syncing subtitles of type Movies")
}
