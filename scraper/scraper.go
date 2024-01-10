// Package scraper provides a function to Scrape UGC info from a JSON file.
//
// A couple of variables can be set to make it function properly. The results are saved in a JSON file or XLSX file according according to the variable resultFormat.
package scraper

import (
	"context"
	"errors"
	"log"
	"net/mail"
	"net/url"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	fileopers "github.com/jcbl1/tiktok_ugc_finder/file_opers"
	ugcinfo "github.com/jcbl1/tiktok_ugc_finder/ugc_info"
	"github.com/jcbl1/tiktok_ugc_finder/utils"
	"golang.org/x/sync/semaphore"
)

// Scrape scrapes UGC info according to a JSON file providing their unique IDs.
//
// It will supposingly save results anyway whether the process has finished successfully or not.
func Scrape(scrapedJSONFile string) error {
	ugcs, err := ugcinfo.FromJSON(scrapedJSONFile) // gets UGCs from JSON.
	if err != nil {
		return err
	}
	if from < 0 || from >= len(ugcs) { // sets from and to to a proper value.
		from = len(ugcs)
	}
	if to < 0 || to >= len(ugcs) {
		to = len(ugcs)
	}
	ugcs = ugcs[from:to] // trims ugcs.
	if len(ugcs) > int(limit) { // respects the limit.
		ugcs = ugcs[:limit]
	}
	log.Println("UGCs to be processed:", len(ugcs))

	errs := make(chan error) // defines a channel to receive errors (if any) in closures.
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt)
	// go func(ugcs *[]ugcinfo.UGCInfo, errChan chan error) {
	// 	for sig := range c {
	// 		log.Println(sig, "detected, saving results")
	// 		if err := saveResults(*ugcs); err != nil {
	// 			errChan <- err
	// 		}
	// 	}
	// }(&ugcs, errs)

	ctx, cancel := context.WithCancel(context.Background()) // defines the main context
	defer cancel()
	defer func(ugcs *[]ugcinfo.UGCInfo, errChan chan error) { // saves to file BEFORE ctx is canceled and this function is done.
		if err := saveResults(*ugcs); err != nil {
			errChan <- err
		}
	}(&ugcs, errs)
	if err := scrapeProfileVideos(ctx, &ugcs); err != nil { // processes ugcs
		return err
	}

	return nil
}

// saveResults uses [fileopers] to save results.
func saveResults(ugcs []ugcinfo.UGCInfo) error {
	log.Println("Saving results")
	switch resultFormat { //respects resultFormat
	case "json":
		if err := fileopers.SaveResultsAsJSON(ugcs); err != nil {
			return err
		}
	case "xlsx":
		if err := fileopers.SaveResultsAsXLSX(ugcs); err != nil {
			return err
		}
	}

	return nil
}

// scrapeProfileVideos does the real job for scraping.
//
// It allocates a browser and simulates the process of navigating, clicking and etc. ctxParent makes it easier to cancel the process when needed. ugcs is passed as a pointer so any changes will immediately take effect on the ugcs in Scrape(). An error is returned if it encounters any error that is due to the function itself (i.e. "Internal Error" is supposed to be returned).
func scrapeProfileVideos(ctxParent context.Context, ugcs *[]ugcinfo.UGCInfo) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false), chromedp.DisableGPU) // customizes options used to allocate a browser.
	var allocCtx context.Context
	var cancel context.CancelFunc
	if headless { // sets allocCtx to the default one if headless is true, or to the customized one if headless is false.
		allocCtx, cancel = chromedp.NewExecAllocator(ctxParent, chromedp.DefaultExecAllocatorOptions[:]...)
	} else {
		allocCtx, cancel = chromedp.NewExecAllocator(ctxParent, opts...)
	}
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf)) // gets the browser context
	defer cancel()

	if err := chromedp.Run( // sets the viewport
		ctx,
		chromedp.EmulateViewport(1280, 720),
	); err != nil {
		return err
	}

	if err := chromedp.Run( // navigates to the first user profile page.
		ctx,
		chromedp.Navigate(TIKTOK+"/@"+(*ugcs)[0].UniqueID),
	); err != nil {
		return err
	}

	var refreshButtons []*cdp.Node // gets refresh button nodes
	if err := chromedp.Run(
		ctx,
		getRefreshButtons(&refreshButtons),
	); err != nil {
		return err
	}
	
	if err := chromedp.Run( // clicks on the refresh button
		ctx,
		chromedp.MouseClickNode(refreshButtons[0], chromedp.ButtonLeft),
	); err != nil {
		return err
	}

	links, err := getProfileVideoLinks(ctx) // gets profile video links
	if err != nil {
		return err
	}

	sem := semaphore.NewWeighted(5) // use semaphore to limit the amount of processes asking API server for help.

	errs := make(chan error, len(*ugcs)) // error channel used to detect errors
	finishes := make(chan int, len(*ugcs)) // channel used to check if all goroutines are done.

	go func(ctx context.Context, errChan chan error, finishChan chan int) { // gets AP and AI
		if verbose {
			log.Println("Getting AP and AI of the first user")
		}
		if lt, ap, ai, err := calculateAPAndAI(ctx, links); err != nil { // calculates AP and AI and if no error, stores them.
			errChan <- err
		} else {
			(*ugcs)[0].AP = ap
			(*ugcs)[0].AI = ai
			(*ugcs)[0].LatestVideoTime = time.Unix(int64(lt), 0)
		}
		finishChan <- 0 // goroutine finished
	}(ctx, errs, finishes)

	// if verbose {
	// 	log.Println("Getting AP and AI of the first user")
	// }
	// if lt, ap, ai, err := calculateAPAndAI(links); err != nil {
	// 	return err
	// } else {
	// 	(*ugcs)[0].AP = ap
	// 	(*ugcs)[0].AI = ai
	// 	(*ugcs)[0].LatestVideoTime = time.Unix(int64(lt), 0)

	// }

	var mails []*mail.Address // gets emails
	if err := findEmails(ctx, &mails); err != nil {
		return err
	}
	for _, m := range mails {
		(*ugcs)[0].Email = append((*ugcs)[0].Email, m.String())
	}

	// Sleep for an hour when testing
	// chromedp.Run(
	// 	ctx,
	// 	chromedp.Sleep(time.Hour),
	// )

	for i := range (*ugcs)[1:] { // for each ugcs left, do the almost same thing as above.
		if err := chromedp.Run(
			ctx,
			chromedp.Navigate(TIKTOK+"/@"+(*ugcs)[i+1].UniqueID),
		); err != nil {
			return err
		}

		links, err := getProfileVideoLinks(ctx)
		if err != nil {
			return err
		}
		log.Printf("Getting AP and AI of the %dth user\n", i+2)
		go func(ctx context.Context, errChan chan error, finishChan chan int, index int) {
			if err := sem.Acquire(context.TODO(), 1); err != nil { // acquires on semaphore
				errChan <- err
			}
			// log.Printf("👻goroutine started[%d]", index)
			if lt, ap, ai, err := calculateAPAndAI(ctx, links); err != nil {
				errChan <- err
			} else {
				// log.Println("👻ap", ap, "ai", ai, "latestVideoTime", time.Unix(int64(lt), 0))
				(*ugcs)[index].AP = ap
				(*ugcs)[index].AI = ai
				(*ugcs)[index].LatestVideoTime = time.Unix(int64(lt), 0)
			}
			// log.Println("👻goroutine finished")
			sem.Release(1) // releases to semaphore
			finishChan <- index
		}(ctx, errs, finishes, i+1)

		if verbose { // gets mails
			log.Println("Getting emails")
		}
		var mails []*mail.Address
		if err := findEmails(ctx, &mails); err != nil {
			return err
		}
		for _, m := range mails {
			(*ugcs)[i+1].Email = append((*ugcs)[i+1].Email, m.String())
		}
	}

	finished := 0 // variable to count how many goroutines are finished.
	for {
		select {
		case err := <-errs:
			return err
		case <-ctx.Done():
			return errors.New("canceled")
		case <-finishes: // finished increments by one and if it equals to the length of ugcs, this function stops waiting and returns nil
			finished++
			// log.Printf("⛳️finishes received[%d]. finished:%d", i, finished)
			if finished == len(*ugcs) {
				return nil
			}
		}
	}
}

// getProfileVideoLinks gets video links that are not pinned on a profile page.
func getProfileVideoLinks(ctx context.Context) ([]string, error) {
	var anchors []*cdp.Node
	var videoCount uint = 0
	var links []string
	if verbose {
		log.Println("Getting links")
	}
	for videoCount < recentVideosNum {
		ctxTimeout, cancel := context.WithTimeout(ctx, time.Second*15)
		defer cancel()

		if err := chromedp.Run(
			ctxTimeout,
			chromedp.Nodes(`//*[@id="main-content-others_homepage"]/div/div[2]/div[*]/div/div[*]/div[1]/div/div/a[not(.//div[@data-e2e="video-card-badge"])]`, &anchors, chromedp.NodeReady),
			scrollToBottom(),
		); err != nil && errors.Is(err, ctxTimeout.Err()) {
			log.Println("😅 refresh")
			// get refresh button node
			var refreshButtons []*cdp.Node
			if err := chromedp.Run(
				ctx,
				getRefreshButtons(&refreshButtons),
			); err != nil {
				return nil, err
			}
			// click on the refresh button
			if err := chromedp.Run(
				ctx,
				chromedp.MouseClickNode(refreshButtons[0], chromedp.ButtonLeft),
			); err != nil {
				return nil, err
			}
		} else if err != nil {
			return nil, err
		}

		// if err := chromedp.Run(
		// 	ctx,
		// 	chromedp.Nodes(`//*[@id="main-content-others_homepage"]/div/div[2]/div[*]/div/div[*]/div[1]/div/div/a[not(.//div[@data-e2e="video-card-badge"])]`, &anchors, chromedp.NodeReady),
		// 	scrollToBottom(),
		// ); err != nil {
		// 	return nil, err
		// }

		//if failed to get nodes, try click refresh button again
		// if len(anchors) == 0 {
		// 	// get refresh button node
		// 	var refreshButtons []*cdp.Node
		// 	if err := chromedp.Run(
		// 		ctx,
		// 		getRefreshButtons(&refreshButtons),
		// 	); err != nil {
		// 		return nil, err
		// 	}
		// 	// click on the refresh button
		// 	if err := chromedp.Run(
		// 		ctx,
		// 		chromedp.MouseClickNode(refreshButtons[0], chromedp.ButtonLeft),
		// 	); err != nil {
		// 		return nil, err
		// 	}
		// }

		links = links[:0]
		//Empty slice now, appending links
		for _, anchor := range anchors {
			link := anchor.AttributeValue("href")
			u, err := url.Parse(link)
			if err != nil {
				return nil, err
			}
			if !(u.Hostname() == "") {
				links = append(links, anchor.AttributeValue("href"))
			}
		}

		if uint(len(anchors)) > videoCount {
			videoCount = uint(len(anchors))
		} else {
			break
		}
	}

	if uint(len(links)) > recentVideosNum {
		links = links[:recentVideosNum]
	}

	return links, nil
}

func scrollToBottom() chromedp.ActionFunc {
	return func(ctx context.Context) error {
		_, exp, err := runtime.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`).Do(ctx)
		if err != nil {
			return err
		}
		if exp != nil {
			return exp
		}
		return nil
	}
}

func debugLog(s string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		log.Println(s)
		return nil
	})
}

// getRefreshButtons returns a task list to get refreshButtons.
func getRefreshButtons(refreshButtons *[]*cdp.Node) chromedp.Tasks {
	return chromedp.Tasks{
		// chromedp.WaitVisible(`//*[@id="login-modal-title"]`),
		chromedp.ActionFunc(func(ctx context.Context) error {
			log.Println("title stuff")
			var titleNodes []*cdp.Node
			if err := chromedp.Run(
				ctx,
				debugLog("waiting body ready"),
				chromedp.WaitReady(`//body`),
				debugLog("sleeping"),
				chromedp.Sleep(utils.MediumInterval()),
				debugLog("getting nodes"),
				chromedp.Nodes(`//*[@id="login-modal-title"]`, &titleNodes, chromedp.AtLeast(0)),
			); err != nil {
				return err
			}
			log.Println("len(titleNodes):", len(titleNodes))
			if len(titleNodes) != 0 {
				if err := chromedp.Run(
					ctx,
					chromedp.KeyEvent(kb.Escape),
				); err != nil {
					return err
				}
			}

			return nil
		}),
		// chromedp.KeyEvent(kb.Escape),
		// chromedp.WaitNotPresent(`//*[@id="login-modal-title"]`),
		chromedp.Sleep(utils.ShortInterval()),
		chromedp.WaitVisible(`//*[@id="main-content-others_homepage"]/div/div[2]/main/div/button`),
		chromedp.Sleep(utils.ShortInterval()),
		chromedp.Nodes(`//*[@id="main-content-others_homepage"]/div/div[2]/main/div/button`, refreshButtons),
	}
}

// calculateAPAndAI trys to get video statistics from API server and will keep trying if it meets errors from other than ctx canceled.
func calculateAPAndAI(ctx context.Context, links []string) (latestVideoTime int, ap int, ai float32, err error) {
	var vss []utils.VideoStats
	for i, link := range links {
		if verbose {
			log.Printf("Getting result of the %dth link: %s", i+1, link)
		}
		var vs utils.VideoStats
		if i == 0 { // means the latest video
			err = backoff.Retry(func() error {
				select {
				case <-ctx.Done():
					return backoff.Permanent(errors.New("ctx canceled"))
				default:
					latestVideoTime, vs, err = utils.GetVideoStatsFromAPI(link)
					if errors.Is(err, utils.ErrAPIBusy) && verbose {
						log.Println("error:", err, "Retrying")
					} else if err != nil {
						log.Println("error:", err, "Retrying")
					}
					return err
				}
			}, backoff.NewExponentialBackOff())
			if err != nil {
				return
			}
		} else {
			err = backoff.Retry(func() error {
				select {
				case <-ctx.Done():
					return backoff.Permanent(errors.New("ctx canceled"))
				default:
					_, vs, err = utils.GetVideoStatsFromAPI(link)
					if errors.Is(err, utils.ErrAPIBusy) && verbose {
						log.Println("error:", err, "Retrying")
					} else if err != nil {
						log.Println("error:", err, "Retrying")
					}
					return err
				}
			}, backoff.NewExponentialBackOff())
			if err != nil {
				return
			}
		}
		vss = append(vss, vs)
	}

	if len(vss) != 0 { // calculation
		// log.Println("👻vss:",vss)
		total := 0
		ai_total := float32(0)
		for _, vs := range vss {
			// log.Println("👻vs:",vs)
			total += vs.PlayCount
			ai_total += float32(vs.DiggCount) / float32(vs.PlayCount)
			// log.Println("👻total:",total,"ai_total:",ai_total)
		}
		ap = total / len(vss)
		// log.Println("👻ap = total/len(vss), ap:",ap,"total:",total,"len(vss)",len(vss))
		ai = ai_total / float32(len(vss))
		// log.Println("👻ai = ai_total/float32(len(vss)), ai:",ai,"ai_total:",ai_total,"float32(len(vss))",float32(len(vss)))
	}
	return
}

// findEmails finds mail on the profile page.
func findEmails(ctx context.Context, mails *[]*mail.Address) error {
	var bodyText string
	if err := chromedp.Run(
		ctx,
		chromedp.Text(`//body`, &bodyText, chromedp.NodeReady),
	); err != nil {
		return err
	}
	if err := utils.FindMails(bodyText, mails); err != nil {
		return err
	}

	return nil
}
