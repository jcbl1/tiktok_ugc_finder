package utils

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
)

func TestGetVideoStatsFromAPI(t *testing.T) {
	link := "https://www.tiktok.com/@fer.faceyoga/video/7310293679493614853"
	a, err := url.Parse("http://127.0.0.1:8000")
	if err != nil {
		t.Error(err)
	}
	apiServer = *a
	var createdTime int
	var vs VideoStats
	if err := backoff.Retry(func() error {
		createdTime, vs, err = GetVideoStatsFromAPI(link)
		if err != nil {
			fmt.Println(err)
		}
		return err
	}, backoff.NewExponentialBackOff()); err != nil {
		t.Error(err)
	}

	fmt.Println(time.Unix(int64(createdTime), 0))
	fmt.Println(vs)
	t.Fail()
}
