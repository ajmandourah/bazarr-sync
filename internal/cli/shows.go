/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/ajmandourah/bazarr-sync/internal/config"
	"github.com/spf13/cobra"
	"os"
	"fmt"
	"strconv"
	"time"

	"github.com/pterm/pterm"

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
	for i, show := range shows.Data {
		episodes, err := bazarr.QueryEpisodes(cfg,show.SonarrSeriesId)
		if err != nil {
			continue
		}

		for _, episode := range episodes.Data {
			for _, subtitle := range episode.Subtitles {
				p,_ := pterm.DefaultSpinner.Start(pterm.LightBlue(show.Title),
					pterm.LightGreen(":",episode.Title),
					" lang:" + pterm.LightRed(subtitle.Code2) + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(shows.Data)))

				if subtitle.Path == "" || subtitle.File_size == 0 {
					pterm.Success.Prefix = pterm.Prefix{Text: "SKIP", Style: pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack)}
					p.Success(pterm.LightBlue(show.Title,":",episode.Title, "Could not find a subtitle. most likely an embedded. Lang: ",subtitle.Code2))
					pterm.Success.Prefix = pterm.Prefix{Text: "SUCCESS", Style: pterm.NewStyle(pterm.BgGreen, pterm.FgBlack)}
					continue
				}
				params := bazarr.GetSyncParams("episode", episode.SonarrEpisodeId, subtitle)
				if gss {params.Gss = "True"}
				if no_framerate_fix {params.No_framerate_fix = "True"}
				ok := bazarr.Sync(cfg, params)
				if ok {
					
					p.Success("Synced ", show.Title,":", episode.Title, " lang: ", subtitle.Code2)
					continue
				} else {
					for i := 1; i < 2; i++ {
						p,_ := pterm.DefaultSpinner.Start(pterm.LightBlue(show.Title),
							pterm.LightGreen(":",episode.Title),
							" lang:" + pterm.LightRed(subtitle.Code2) + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(shows.Data)))
						time.Sleep(2 * time.Second)
						ok := bazarr.Sync(cfg, params)
						if ok {
							p.Success("Synced ", show.Title,":", episode.Title, " lang: ", subtitle.Code2)
							break
						}
					}
					if !ok{
						p.Fail("Unable to sync ", show.Title, ":", episode.Title, " lang: ", subtitle.Code2)
					}
				}
			}
		}
	}
	fmt.Println("Finished syncing subtitles of type Movies")
}
