/*
Copyright © 2024 ajmandourah
*/
package cli

import (
	"os"
	"github.com/ajmandourah/bazarr-sync/internal/config"

	"github.com/spf13/cobra"
)

var gss bool
var no_framerate_fix bool
var to_list bool 
// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "bazarr-sync",
	Example: "bazarr-sync --config config.yaml sync movies --no-framerate-fix",
	Short: "Manually bulk-sync subtitles downloaded via bazarr",
	Long: `Bulk-sync subtitles downloaded via Bazarr.

Bazarr let you download subs for your titles automatically. 
But if for some reason you needed to sync old subtitles, wither you need to do it because you have not synced them before or you have edited them elsewhere, you will be forced to do it one by one as there is no option to bulk sync them.

This cli tool help you achieve that by utilizing bazarr's api. 

Make sure to create a config.yaml file including your configuration in it. Use the provided config file as a template.
	`, 

}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize( func () {
	
		config.InitConfig()
	})

	rootCmd.PersistentFlags().StringVar(&config.CfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.PersistentFlags().BoolVar(&gss,"golden-section",false,"Use Golden-Section Search")
	rootCmd.PersistentFlags().BoolVar(&no_framerate_fix,"no-framerate-fix",false,"Don't try to fix framerate")

	rootCmd.PersistentFlags().BoolVar(&to_list,"list",false,"list your media with their respective imdbId")
}

