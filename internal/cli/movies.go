/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"github.com/ajmandourah/bazarr-sync/internal/config"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/ajmandourah/bazarr-sync/internal/bazarr"
	"github.com/pterm/pterm"

	"github.com/spf13/cobra"
)

var radarrid []int

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
		if to_list {
			list_movies(cfg)
			return
		}
		sync_movies(cfg)
	},
}

func init() {
	syncCmd.AddCommand(moviesCmd)
	
	moviesCmd.Flags().IntSliceVar(&radarrid,"radarr-id",[]int{},"Specify a list of radarr Ids to sync. Use --list to view your movies with respective radarr id.")
}

func sync_movies(cfg config.Config) {
	movies, err := bazarr.QueryMovies(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr,"Query Error: Could not query movies")
	}
	fmt.Println("Syncing Movies in your Bazarr library.")
	
	movies:
	for i, movie := range movies.Data {
		if len(radarrid) > 0 {
			for _, id := range radarrid {
				if id == movie.RadarrId {
					goto subtitle
				}
			}
			continue movies
		}

		subtitle:
		for _,subtitle := range movie.Subtitles {
			p,_ := pterm.DefaultSpinner.Start(pterm.LightBlue(movie.Title) + " lang:" + pterm.LightRed(subtitle.Code2) + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(movies.Data)))
			if subtitle.Path == "" || subtitle.File_size == 0 {
				pterm.Success.Prefix = pterm.Prefix{Text: "SKIP", Style: pterm.NewStyle(pterm.BgLightBlue, pterm.FgBlack)}
				p.Success(pterm.LightBlue(movie.Title," Could not find a subtitle. most likely it is embedded. Lang: ",subtitle.Code2))
				pterm.Success.Prefix = pterm.Prefix{Text: "SUCCESS", Style: pterm.NewStyle(pterm.BgGreen, pterm.FgBlack)}
				continue
			}
			params := bazarr.GetSyncParams("movie", movie.RadarrId, subtitle)
			if gss {params.Gss = "True"}
			if no_framerate_fix {params.No_framerate_fix = "True"}
			ok := bazarr.Sync(cfg,params)	
			if ok {
				p.Success("Synced ", movie.Title, " lang:", subtitle.Code2)
				continue

			} else {
				p.Warning("Error while syncing ", movie.Title, " lang: ", subtitle.Code2, " Retrying..")
				for i := 1; i < 2; i++{	
					p,_ := pterm.DefaultSpinner.Start(pterm.LightBlue(movie.Title) + " lang:" + pterm.LightRed(subtitle.Code2) + " " + strconv.Itoa(i+1) + "/" + strconv.Itoa(len(movie.Subtitles)))
					time.Sleep(2*time.Second)
					ok := bazarr.Sync(cfg, params)
					if ok{
						p.Success("Synced: ", movie.Title, " lang:", subtitle.Code2)
						break
					}	
				}
				if !ok {	
					p.Fail("Unable to sync ", movie.Title, " lang: ", subtitle.Code2)
				}
			}
			
		}
	} 
	fmt.Println("Finished syncing subtitles of type Movies")
}

func list_movies(cfg config.Config) {	
	movies, err := bazarr.QueryMovies(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr,"Query Error: Could not query movies")
	}
	table := pterm.TableData{
		{"Title","RadarrId"},
	}
	pterm.Println(pterm.LightGreen("Listing all your movies with their respective imdbId (great for syncing specefic movie)\n"))

	for _, movie := range movies.Data {
		// pterm.Println(pterm.LightBlue(movie.Title), "\t", pterm.LightRed(movie.ImdbId))
		t := []string{pterm.LightBlue(movie.Title),pterm.LightRed(movie.RadarrId)}
		table = append(table, t)
	}
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(table).Render()
}
