// Package cmd implements CLI related functions and variables for tiktok_ugc_finder. The expected entry is rootCmd.Execute(), which parses flags and arguments and execute the command.
package cmd

import (
	"log"
	"net/url"
	"path"

	fileopers "github.com/jcbl1/tiktok_ugc_finder/file_opers"
	"github.com/jcbl1/tiktok_ugc_finder/scraper"
	ugcinfo "github.com/jcbl1/tiktok_ugc_finder/ugc_info"
	"github.com/jcbl1/tiktok_ugc_finder/utils"
	"github.com/spf13/cobra"
)

// The root command executed when no subcommands are specified.
var rootCmd = &cobra.Command{
	Run: root,
	Use: "tiktok_ugc_finder",
}

// Variables to store flags.
var (
	recentVideosNum                    uint
	workingDir                         string
	minFollowerCount, maxFollowerCount string
	scrapedJSONFile                    string
	apiServer                          string
	resultFormat                       string
	verbose                            bool
	limit                              uint
	headless                           bool
	from                               int
	to                                 int
)

// init defines all custom flags that can be parsed by root command.
func init() {
	rootCmd.Flags().UintVarP(&recentVideosNum, "recent-videos-num", "R", 15, "Number of videos counted when calculating average-plays (AP) and average interactionality (AI)")
	rootCmd.Flags().StringVarP(&workingDir, "working-dir", "d", ".", "Working directory to store screenshots, tmp files, excel outputs and etc.")
	rootCmd.Flags().StringVarP(&minFollowerCount, "min-follower-count", "m", "0", "Minimum follower count to be selected, in unit K (thousand), M (million)")
	rootCmd.Flags().StringVarP(&maxFollowerCount, "max-follower-count", "M", "INF", "Maximum follower count to be selected, in unit K (thousand), M (million)")
	rootCmd.Flags().StringVarP(&scrapedJSONFile, "scraped-json-file", "j", "", "Scraped JSON file to be processed")
	rootCmd.Flags().StringVarP(&apiServer, "api-server", "A", "http://127.0.0.1:8000", "API server used to get video info from link")
	rootCmd.Flags().StringVarP(&resultFormat, "result-format", "F", "json", "file format to save results (json/xlsx/xml/toml/yml)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "More detailed logs")
	rootCmd.Flags().UintVar(&limit, "limit", 10086, "Limit number of UGCs")
	rootCmd.Flags().BoolVar(&headless, "headless", false, "Whether to use headless mode")
	rootCmd.Flags().IntVar(&from, "from", 0, "From which ugc (by indexing starting from 0) the scraper should process (inclusive). Negative numbers are considered as the total number of unique ugcs")
	rootCmd.Flags().IntVar(&to, "to", -1, "To which ugc (by indexing starting from 0) the scraper should process (exclusive). Negative numbers are considered as the total number of unique ugcs")
}

// root is the actual endpoint where rootCmd is executed.
//
// It sets a bunch of variables in related packages and calls scraper.Scrape() to do further stuff.
func root(cmd *cobra.Command, args []string) {
	if verbose {
		log.Println("verbose mode")
	}
	fileopers.SetWorkingDir(path.Clean(workingDir)) // sets working directory used by fileopers
	scraper.SetVerbose(verbose) // sets verbose mode for [scraper]
	scraper.SetRecentVideosNum(recentVideosNum)
	scraper.SetResultFormat(resultFormat)
	scraper.SetLimit(limit)
	scraper.SetHeadless(headless)
	scraper.SetFromTo(from, to)
	ugcinfo.SetVerbose(verbose) //sets verbose mode for [ugcinfo]
	if err := ugcinfo.SetMinMaxFollowerCount(minFollowerCount, maxFollowerCount); err != nil { // sets minFollowerCount and maxFollowerCount for ugcinfo and crashes on error.
		log.Fatalln(err)
	}
	as, err := url.Parse(apiServer) // parses API server URL.
	if err != nil {
		log.Fatalln(err)
	}
	utils.SetAPIServer(as) // sets API server used by [utils]
	if err := scraper.Scrape(path.Clean(scrapedJSONFile)); err != nil { // starts the scraping process, watching for errors.
		log.Fatalln(err)
	}
}

// Execute is the entry of rootCmd where binary packages can use.
func Execute() {
	rootCmd.Execute()
}
