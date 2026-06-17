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

var radarrid []int
var moviesContinueFrom int

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
		fmt.Fprintln(os.Stderr, "[DEBUG] Using ComputedBaseUrl:", cfg.ComputedBaseUrl)
		fmt.Fprintln(os.Stderr, "[DEBUG] Using ApiUrl:", cfg.ApiUrl)
		fmt.Fprintln(os.Stderr, "[DEBUG] Using Token prefix:", cfg.ApiToken[:4]+"...")
		bazarr.HealthCheck(cfg)
		if to_list {
			list_movies(cfg)
			return
		}
		runWithSignalHandler(func(c chan int){
			sync_movies(cfg, c)
		})
	},
}

func init() {
	syncCmd.AddCommand(moviesCmd)
	
	moviesCmd.Flags().IntSliceVar(&radarrid,"radarr-id",[]int{},"Specify a list of radarr Ids to sync. Use --list to view your movies with respective radarr id.")
	moviesCmd.Flags().IntVar(&moviesContinueFrom,"continue-from",-1,"Continue with the given Radarr movie ID.")
}

// printMovieStatus prints a status line to the correct stream.
// In TTY mode: writes to stderr (same as spinner). In non-TTY: writes to stdout.
func printMovieStatus(isTerminal bool, prefix string, message string) {
	if isTerminal {
		fmt.Fprint(os.Stderr, "\033[K") // clear rest of spinner line
	}
	if prefix != "" {
		fmt.Printf("  %s", prefix)
	}
	if message != "" {
		fmt.Println(message)
	} else {
		fmt.Println()
	}
}

func sync_movies(cfg config.Config, c chan int) {
	movies, err := bazarr.QueryMovies(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query movies:", err)
		os.Exit(1)
	}
	fmt.Println("Syncing Movies in your Bazarr library.")

	skipForward := moviesContinueFrom != -1
	stats := syncStats{}
	isTerminal := term.IsTerminal(int(os.Stdout.Fd()))
moviesLoop:
	for _, movie := range movies.Data {
		if len(radarrid) > 0 {
			for _, id := range radarrid {
				if id == movie.RadarrId {
					goto subtitle
				}
			}
			continue moviesLoop
		}

		if skipForward {
			if movie.RadarrId == moviesContinueFrom {
				skipForward = false
			} else {
				s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
				s.Writer = os.Stderr // stderr — never corrupt stdout on narrow/resize
				s.Start()
				printMovieStatus(isTerminal, "SKIP", "(continue)")
				stats.skipped++
				continue
			}
		}

subtitle:
		c <- movie.RadarrId
		for _, subtitle := range movie.Subtitles {
			if isTerminal {
				s := spinner.New(spinner.CharSets[39], 100*time.Millisecond)
				s.Writer = os.Stderr // stderr — never corrupt stdout on narrow/resize
				s.Start()
				c <- movie.RadarrId
				if subtitle.Path == "" || subtitle.File_size == 0 {
					printMovieStatus(isTerminal, "SKIP", "(no sub file)")
					stats.skipped++
					continue
				}
				params := bazarr.GetSyncParams("movie", movie.RadarrId, subtitle)
				if gss {params.Gss = "True"}
				if no_framerate_fix {params.No_framerate_fix = "True"}

				ok := bazarr.Sync(cfg, params)
				if ok {
					printMovieStatus(isTerminal, "SYNCED", "")
					stats.success++
					continue
				}

				// Retry with exponential backoff
				fmt.Fprint(os.Stderr, "\033[K") // clear spinner line
				fmt.Fprintf(os.Stderr, "  WARNING: %s lang:%s\n", "Error while syncing", subtitle.Code2)
				ok = retrySync(cfg, params, movie.Title, subtitle.Code2)
				if ok {
					printMovieStatus(isTerminal, "SYNCED", "")
					stats.success++
				} else {	
					printMovieStatus(isTerminal, "FAILED", "")
					stats.failed++
				}
			} else {
				// Non-TTY: simple text output, no spinner animation
				if subtitle.Path == "" || subtitle.File_size == 0 {
					fmt.Printf("  SKIP %s lang=%s (no subtitle file)\n", movie.Title, subtitle.Code2)
					stats.skipped++
					continue
				}
				params := bazarr.GetSyncParams("movie", movie.RadarrId, subtitle)
				if gss {params.Gss = "True"}
				if no_framerate_fix {params.No_framerate_fix = "True"}

				ok := bazarr.Sync(cfg, params)
				if ok {
					fmt.Printf("  SYNCED %s lang=%s\n", movie.Title, subtitle.Code2)
					stats.success++
					continue
				} else {	
					fmt.Printf("  FAILED %s lang=%s\n", movie.Title, subtitle.Code2)
					stats.failed++
				}
			}
		}
	}
	fmt.Println("Finished syncing subtitles of type Movies")
	fmt.Printf("\n📊 Summary: %d synced, %d skipped, %d failed\n", stats.success, stats.skipped, stats.failed)
	// Signal that we're done with all subtitles.
	close(c)
}

func list_movies(cfg config.Config) {
	movies, err := bazarr.QueryMovies(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Query Error: Could not query movies:", err)
		os.Exit(1)
	}
	table := pterm.TableData{
		{"Title","RadarrId"},
	}
	pterm.Println(pterm.LightGreen("Listing all your movies with their respective Radarr ID (great for syncing specific movies)\n"))

	for _, movie := range movies.Data {
		t := []string{pterm.LightBlue(movie.Title),pterm.LightRed(movie.RadarrId)}
		table = append(table, t)
	}
	pterm.DefaultTable.WithHasHeader().WithHeaderRowSeparator("-").WithData(table).Render()
}
