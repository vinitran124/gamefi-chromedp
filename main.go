package main

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"log"
	"sync"
	"time"
)

const (
	maxGoroutines    = 4
	amountRepetition = 10_000_000

	gamefiUrl    = "https://gamefi.org/"
	expandBtn    = `//*[@id="radix-0"]`
	launchpadBtn = `//*[@id="radix-1"]/div[1]/a`
	inoBtn       = `//*[@id="gameDetailContent"]/div[2]/a[2]`
	gameWorldBtn = `//*[@id="gameDetailContent"]/div[2]/a[3]`
	insightBtn   = `//*[@id="gameDetailContent"]/div[3]/div[2]/a[7]`

	launchpadItemBtn = `//*[@id="layoutBody"]/main/div[2]/div[2]/div/div/div[%d]/div/a[1]`
	inoItemBtn       = `//*[@id="layoutBody"]/main/div[3]/div[%d]/a`
	gameWorldItemBtn = `//*[@id="layoutBody"]/main/div[2]/div/div[2]/div/div[3]/div/div/div[%d]/a`
	insightItemBtn   = `//*[@id="layoutBody"]/main/div/div[3]/div[%d]/a`

	expandBottomBtn = `//*[@id="gameDetailContent"]/div[2]/button`
)

func main() {
	newChromedp()
}

func newChromedp() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("start-fullscreen", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", false),
		chromedp.Flag("remote-debugging-port", "9222"),
		chromedp.WindowSize(480, 1080),
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
	)
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	if err := chromedp.Run(ctx,
		fetch.Enable()); err != nil {
		log.Println(err)
		return
	}

	guard := make(chan struct{}, maxGoroutines)
	waitGroup := sync.WaitGroup{}

	for i := 0; i < amountRepetition; i++ {
		guard <- struct{}{}
		waitGroup.Add(1)
		go func(i int) {
			windownFeature := fmt.Sprintf("scrollbars,status,width=%d, height=%d, top=%d, left=%d", 480, 1080, 0, 480*(i%4))
			windown := fmt.Sprintf(`window.open("about:blank", "", "%s")`, windownFeature)
			err := NewWindown(ctx, windown)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(25 * time.Second)
			<-guard
			waitGroup.Done()
		}(i)
	}
	waitGroup.Wait()
}

func NewWindown(ctx context.Context, windown string) error {
	var res *runtime.RemoteObject
	if err := chromedp.Run(ctx, chromedp.Evaluate(windown, &res)); err != nil {
		return err
	}

	targets, err := chromedp.Targets(ctx)
	if err != nil {
		return err
	}

	for _, t := range targets {
		if !t.Attached {
			ctx, _ := context.WithTimeout(ctx, 4*time.Minute)
			newCtx, cancel := chromedp.NewContext(ctx, chromedp.WithTargetID(t.TargetID))

			if err := chromedp.Run(newCtx, GamefiTask()); err != nil {
				return err
			}
			cancel()
			break
		}
	}

	return nil
}

func GamefiTask() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(gamefiUrl),
		chromedp.Sleep(2 * time.Second),
		chromedp.Click(expandBtn),
		chromedp.Sleep(2 * time.Second),
		AccessTab(launchpadBtn, launchpadItemBtn),
		AccessTab(inoBtn, inoItemBtn),
		AccessTab(gameWorldBtn, gameWorldItemBtn),
		AccessInsight(),
		chromedp.Stop(),
	}
}

func ScrollWeb() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.KeyEvent(kb.PageDown),
		chromedp.Sleep(3 * time.Second),
		chromedp.KeyEvent(kb.PageDown),
		chromedp.Sleep(3 * time.Second),
		chromedp.KeyEvent(kb.PageDown),
		chromedp.Sleep(3 * time.Second),
		chromedp.KeyEvent(kb.PageDown),
		chromedp.Sleep(3 * time.Second),

		chromedp.KeyEvent(kb.PageUp),
		chromedp.Sleep(3 * time.Second),
		chromedp.KeyEvent(kb.PageUp),
		chromedp.Sleep(3 * time.Second),
		chromedp.KeyEvent(kb.PageUp),
		chromedp.Sleep(3 * time.Second),
		chromedp.KeyEvent(kb.PageUp),
		chromedp.Sleep(3 * time.Second),
	}
}

func AccessInsight() chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			if err := chromedp.Run(ctx,
				chromedp.Click(expandBottomBtn),
				chromedp.Sleep(2*time.Second),
				chromedp.Click(insightBtn),
				chromedp.Sleep(2*time.Second),
			); err != nil {
				return err
			}

			if err := chromedp.Run(ctx, ScrollWeb(), chromedp.Sleep(2*time.Second)); err != nil {
				return err
			}

			for i := 1; i < 3; i++ {
				itemBtn := fmt.Sprintf(insightItemBtn, i)
				if err := chromedp.Run(ctx,
					chromedp.Click(itemBtn),
					ScrollWeb(),
					chromedp.NavigateBack(),
					chromedp.Sleep(2*time.Second),
				); err != nil {
					return err
				}
			}
			return nil
		}),
	}
}

func AccessTab(xPathBtn, xPathItemBtn string) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.ActionFunc(func(ctx context.Context) error {
			if err := chromedp.Run(ctx,
				chromedp.Click(xPathBtn),
				chromedp.Sleep(3*time.Second),
			); err != nil {
				return err
			}

			if err := chromedp.Run(ctx, ScrollWeb(), chromedp.Sleep(2*time.Second)); err != nil {
				return err
			}

			for i := 2; i < 3; i++ {
				itemBtn := fmt.Sprintf(xPathItemBtn, i)
				if err := chromedp.Run(ctx,
					chromedp.Click(itemBtn),
					ScrollWeb(),
					chromedp.NavigateBack(),
					chromedp.Sleep(2*time.Second),
				); err != nil {
					return err
				}
			}

			return nil
		}),
	}
}
