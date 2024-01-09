package utils

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
)

var apiServer url.URL

type VideoStats struct {
	DiggCount int `json:"digg_count"`
	PlayCount int `json:"play_count"`
}

type APIResult struct {
	CreateTime int        `json:"create_time"`
	Statistics VideoStats `json:"statistics"`
}

func SetAPIServer(a *url.URL) {
	apiServer = *a
	if apiServer.Scheme == "" {
		apiServer.Scheme = "http"
	}
}

func GetVideoStatsFromAPI(url string) (createdTime int, vs VideoStats, err error) {
	prefix := apiServer.Scheme + "://" + apiServer.Hostname() + ":" + apiServer.Port() + "/api?url="
	// fmt.Println(prefix + url)
	resp, err := http.Get(prefix + url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var res APIResult
	if err = json.Unmarshal(data, &res); err != nil {
		return
	}
	// if api returns empty result, throw an error
	var empty APIResult
	if res == empty {
		err = ErrAPIBusy
		return
	}
	createdTime = res.CreateTime
	// log.Println("👻GetVideoStatsFromAPI: res.Statistics:",res.Statistics)
	return createdTime, res.Statistics, nil
}

var ErrAPIBusy = errors.New("api busy")
