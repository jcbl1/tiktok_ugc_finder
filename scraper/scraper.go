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

func Scrape(scrapedJSONFile string) error {
	ugcs, err := ugcinfo.FromJSON(scrapedJSONFile)
	if err != nil {
		return err
	}
	if from < 0 || from >= len(ugcs) {
		from = len(ugcs)
	}
	if to < 0 || to >= len(ugcs) {
		to = len(ugcs)
	}
	ugcs = ugcs[from:to]
	if len(ugcs) > int(limit) {
		ugcs = ugcs[:limit]
	}
	log.Println("UGCs to be processed:", len(ugcs))

	errs := make(chan error)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func(ugcs *[]ugcinfo.UGCInfo, errChan chan error) {
		if err := saveResults(*ugcs); err != nil {
			errChan <- err
		}
	}(&ugcs, errs)
	if err := scrapeProfileVideos(ctx, &ugcs); err != nil {
		return err
	}

	if err := saveResults(ugcs); err != nil {
		return err
	}

	return nil
}
func saveResults(ugcs []ugcinfo.UGCInfo) error {
	log.Println("Saving results")
	switch resultFormat {
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

func scrapeProfileVideos(ctxParent context.Context, ugcs *[]ugcinfo.UGCInfo) error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", false), chromedp.DisableGPU)
	var allocCtx context.Context
	var cancel context.CancelFunc
	if headless {
		allocCtx, cancel = chromedp.NewExecAllocator(ctxParent, chromedp.DefaultExecAllocatorOptions[:]...)
	} else {
		allocCtx, cancel = chromedp.NewExecAllocator(ctxParent, opts...)
	}
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	// Set viewport
	if err := chromedp.Run(
		ctx,
		chromedp.EmulateViewport(1280, 720),
	); err != nil {
		return err
	}

	if err := chromedp.Run(
		ctx,
		chromedp.Navigate(TIKTOK+"/@"+(*ugcs)[0].UniqueID),
	); err != nil {
		return err
	}

	// get refresh button node
	var refreshButtons []*cdp.Node
	if err := chromedp.Run(
		ctx,
		getRefreshButtons(&refreshButtons),
	); err != nil {
		return err
	}
	// click on the refresh button
	if err := chromedp.Run(
		ctx,
		chromedp.MouseClickNode(refreshButtons[0], chromedp.ButtonLeft),
	); err != nil {
		return err
	}

	// get profile video links
	links, err := getProfileVideoLinks(ctx)
	if err != nil {
		return err
	}

	sem := semaphore.NewWeighted(5)

	errs := make(chan error, len(*ugcs))
	finishes := make(chan int, 2*len(*ugcs))
	// get ap ai
	go func(ctx context.Context, errChan chan error, finishChan chan int) {
		if verbose {
			log.Println("Getting AP and AI of the first user")
		}
		if lt, ap, ai, err := calculateAPAndAI(ctx, links); err != nil {
			errChan <- err
		} else {
			(*ugcs)[0].AP = ap
			(*ugcs)[0].AI = ai
			(*ugcs)[0].LatestVideoTime = time.Unix(int64(lt), 0)
		}
		finishChan <- 0
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

	// get emails
	var mails []*mail.Address
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

	for i := range (*ugcs)[1:] {
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
			if err := sem.Acquire(context.TODO(), 1); err != nil {
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
			sem.Release(1)
			finishChan <- index
		}(ctx, errs, finishes, i+1)

		// get emails
		if verbose {
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

	finished := 0
	for {
		select {
		case err := <-errs:
			return err
		case <-ctx.Done():
			return errors.New("canceled")
		case <-finishes:
			finished++
			// log.Printf("⛳️finishes received[%d]. finished:%d", i, finished)
			if finished == len(*ugcs) {
				return nil
			}
		}
	}
}

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

func calculateAPAndAI(ctx context.Context, links []string) (latestVideoTime int, ap int, ai float32, err error) {
	var vss []utils.VideoStats
	for i, link := range links {
		if verbose {
			log.Printf("Getting result of the %dth link: %s", i+1, link)
		}
		var vs utils.VideoStats
		if i == 0 {
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

	if len(vss) != 0 {
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
