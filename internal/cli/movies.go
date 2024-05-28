/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"ajmandourah/bazarr-sync/internal/bazarr"
	"ajmandourah/bazarr-sync/internal/config"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

// moviesCmd represents the movies command
var moviesCmd = &cobra.Command{
	Use:   "movies",
	Short: "Sync subtitles to the audio track of the movie",
	Example: "bazarr-sync --config config.yaml sync movies --no-framerate-fix",
	Long: `By default, Bazarr will try to sync the sub to the audio track:0 of the media. 
This can fail due to many reasons mainly due to failure of bazarr to extract audio info. This is unfortunatly out of my hands.
The script by default will try to not use the golden section search method and will try to fix framerate issues. This can be changed using the flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.GetConfig()
		bazarr.HealthCheck(cfg)
		sync_movies(cfg)
	},
}

func init() {
	syncCmd.AddCommand(moviesCmd)
}

func sync_movies(cfg config.Config) {
	movies, err := bazarr.QueryMovies(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr,"Query Error: Could not query movies")
	}
	fmt.Println("Syncing ", len(movies.Data), " Movies in your Bazarr library.")

	bar := progressbar.New(len(movies.Data))
	bar.Describe("Syncing..")
	for _, movie := range movies.Data {
		bar.Add(1)
		bar.Describe(movie.Title)
		for i,subtitle := range movie.Subtitles {
			if subtitle.Path == "" || subtitle.File_size == 0 {
				continue
			}
			bar.Describe(movie.Title + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(movie.Subtitles)))
			params := bazarr.GetSyncParams("movie", movie.RadarrId, subtitle)
			if gss {params.Gss = "True"}
			if no_framerate_fix {params.No_framerate_fix = "True"}
			ok := bazarr.Sync(cfg,params)	
			if !ok {
				for i := 1; i < 5; i++{
					time.Sleep(2*time.Second)
					ok := bazarr.Sync(cfg, params)
					if ok{
						break
					}	
				}
				bar.Clear()
				fmt.Println("Unable to sync subtitle for", movie.Title, " lang: ", subtitle.Code2)
				bar.RenderBlank()
			}
		}
	} 
	bar.Clear()
	fmt.Println("Finished syncing subtitles of type Movies")
}