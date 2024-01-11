package cmd

import (
	"errors"
	"log"
	"net/url"
	"path"
	"strings"

	fileopers "github.com/jcbl1/tiktok_ugc_finder/file_opers"
	"github.com/jcbl1/tiktok_ugc_finder/scraper"
	ugcinfo "github.com/jcbl1/tiktok_ugc_finder/ugc_info"
	"github.com/jcbl1/tiktok_ugc_finder/utils"
	"github.com/spf13/cobra"
)

var (
	filename string
)

var mendCmd = &cobra.Command{
	Use: "mend <file>",
	Run: mend,
}

func mend(cmd *cobra.Command, args []string) {
	switch len(args) {
	case 0:
		cmd.Help()
		return
	case 1:
		filename = args[0]
	default:
		cmd.PrintErrln(errors.New("unrecognizable args: " + strings.Join(args[1:], " ")))
		cmd.Help()
		return
	}

	if verbose {
		log.Println("verbose mode")
		log.Println("filename:", filename)
	}
	fileopers.SetVerbose(verbose)
	fileopers.SetWorkingDir(path.Dir(filename))
	scraper.SetVerbose(verbose)
	scraper.SetRecentVideosNum(recentVideosNum)
	switch strings.ToLower(strings.Split(filename, ".")[len(strings.Split(filename, ","))-1]) {
	case "xlsx":
		scraper.SetResultFormat("xlsx")
	case "json":
		scraper.SetResultFormat("json")
	default:
		log.Fatalln(errors.New("file format not supported"))
	}
	scraper.SetHeadless(headless)
	ugcinfo.SetVerbose(verbose)
	// if err := ugcinfo.SetMinMaxFollowerCount(minFollowerCount, maxFollowerCount); err != nil {
	// 	log.Fatalln(err)
	// }
	as, err := url.Parse(apiServer)
	if err != nil {
		log.Fatalln(err)
	}
	utils.SetAPIServer(as)

	if err := scraper.ScrapeUnscraped(filename); err != nil {
		log.Fatalln(err)
	}
}
