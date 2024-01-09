package scraper

import (
	"context"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	fileopers "github.com/jcbl1/tiktok_ugc_finder/file_opers"
)

func Search(hashtags string) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.DisableGPU, chromedp.Flag("headless", false))
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Search by hashtags (actually by queries)
	var screenshotOfSearchPage []byte
	if err := chromedp.Run(
		ctx,
		chromedp.Navigate(TIKTOK_SEARCH+hashtags),
		chromedp.WaitVisible(`//body`),
		chromedp.Sleep(time.Second),
		chromedp.FullScreenshot(&screenshotOfSearchPage, 100),
	); err != nil {
		return err
	}
	if err := fileopers.SaveScreenshot(screenshotOfSearchPage); err != nil {
		return err
	}

	return nil
}
