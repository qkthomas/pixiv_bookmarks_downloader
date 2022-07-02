// MIT License

// Copyright (c) [2022] [Lin Chen]

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package common

import (
	"context"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

func StartSavingResponseToFile(wg *sync.WaitGroup, ctx context.Context, requestID network.RequestID, filepath string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(config.SavingRespWaitDura) //sleep for trying to avoid incomplete images
		param := network.GetResponseBody(requestID)
		if param == nil {
			return
		}
		c := chromedp.FromContext(ctx)
		buf, err := param.Do(cdp.WithExecutor(ctx, c.Target))
		if err != nil {
			fmt.Printf("error when doing param.Do(ctx): %+v\n", err)
			return
		}
		fmt.Printf("writing to file \"%s\"\n", filepath)
		if err := ioutil.WriteFile(filepath, buf, config.WriteFilePermission); err != nil {
			fmt.Printf("error: failed to write to %s: %+v\n", filepath, err)
			return
		}
		fmt.Printf("wrote %s\n", filepath)
	}()
}

func ListenForNetworkEventAndDownloadImages(ctx context.Context,
	urlMatcher func(string) (string, bool)) (waitFunc func()) {
	wg := new(sync.WaitGroup)
	waitFunc = func() {
		fmt.Println("waiting writing files to be done")
		wg.Wait()
	}

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {

		case *network.EventResponseReceived:
			resp := ev.Response
			if len(resp.Headers) != 0 && string(ev.Type) == `Image` {
				// log.Printf("received \"%s\", requestID: %s, resource url: %s\n", ev.Type, ev.RequestID, resp.URL)
			} else {
				return
			}

			filePath, toDownload := urlMatcher(resp.URL)
			if !toDownload {
				return
			}

			StartSavingResponseToFile(wg, ctx, ev.RequestID, filePath)
		}
		// other needed network Event
	})

	return waitFunc
}
