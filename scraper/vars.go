package scraper

import (
	"log"
)

var (
	recentVideosNum                    uint
	resultFormat                       string
	verbose                            bool
	limit                              uint
	headless                           bool
	minFollowerCount, maxFollowerCount int
	from, to                           int
)

// func SetVars(rvn uint,rf string,v bool,l uint,h bool,m,M string,fr,t int){
// 	recentVideosNum=rvn
// 	resultFormat=rf
// 	verbose=v
// 	limit=l
// 	headless=h
// 	minFollowerCount=m
// 	maxFollowerCount=M
// 	from=fr
// 	to=t
// }

func SetRecentVideosNum(r uint) {
	recentVideosNum = r
	if verbose {
		log.Println("recentVideosNum:", recentVideosNum)
	}
}

func SetResultFormat(f string) {
	resultFormat = f
	if verbose {
		log.Println("resultFormat:", resultFormat)
	}
}

func SetVerbose(v bool) {
	verbose = v
}

func SetLimit(l uint) {
	limit = l
	if verbose {
		log.Println("limit:", limit)
	}
}

func SetHeadless(h bool) {
	headless = h
	if verbose {
		log.Println("headless:", headless)
	}
}

func SetFromTo(f, t int) {
	from = f
	to = t
	if verbose {
		log.Println("from", from, "to", to)
	}
}
