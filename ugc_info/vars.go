package ugcinfo

import (
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	verbose                            bool
	minFollowerCount, maxFollowerCount int
)

func SetVerbose(v bool) {
	verbose = v
}

func SetMinMaxFollowerCount(m, M string) error {
	re := regexp.MustCompile(`[0-9]+`)
	minStr := re.FindString(m)
	min, err := strconv.Atoi(minStr)
	if err != nil {
		return err
	}
	switch strings.ToLower(strings.TrimPrefix(m, minStr)) {
	case "k":
		minFollowerCount = min * 1_000
	case "m":
		minFollowerCount = min * 1_000_000
	default:
		minFollowerCount = min
	}

	if M != "INF" {
		maxStr := re.FindString(M)
		max, err := strconv.Atoi(maxStr)
		if err != nil {
			return err
		}
		switch strings.ToLower(strings.TrimPrefix(M, maxStr)) {
		case "k":
			maxFollowerCount = max * 1_000
		case "m":
			maxFollowerCount = max * 1_000_000
		default:
			maxFollowerCount = max
		}
	} else {
		maxFollowerCount = math.MaxInt
	}

	if verbose {
		log.Println("minFollowerCount", minFollowerCount, "maxFollowerCount", maxFollowerCount)
	}

	return nil
}
