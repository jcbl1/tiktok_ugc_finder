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

var rootCmd = &cobra.Command{
	Run: root,
	Use: "tiktok_ugc_finder",
}

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

func root(cmd *cobra.Command, args []string) {
	if verbose {
		log.Println("verbose mode")
	}
	fileopers.SetWorkingDir(path.Clean(workingDir))
	scraper.SetVerbose(verbose)
	scraper.SetRecentVideosNum(recentVideosNum)
	scraper.SetResultFormat(resultFormat)
	scraper.SetLimit(limit)
	scraper.SetHeadless(headless)
	scraper.SetFromTo(from, to)
	ugcinfo.SetVerbose(verbose)
	if err := ugcinfo.SetMinMaxFollowerCount(minFollowerCount, maxFollowerCount); err != nil {
		log.Fatalln(err)
	}
	as, err := url.Parse(apiServer)
	if err != nil {
		log.Fatalln(err)
	}
	utils.SetAPIServer(as)
	if err := scraper.Scrape(path.Clean(scrapedJSONFile)); err != nil {
		log.Fatalln(err)
	}
}

func Execute() {
	rootCmd.Execute()
}
