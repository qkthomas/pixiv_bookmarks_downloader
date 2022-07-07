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
package download

import (
	"context"
	"fmt"
	"path"
	"sync"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/common"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

func ListenForNetworkEventAndDownloadBookmarkThumbnails(ctx context.Context) (waitFunc func() error) {
	urlMatcher := func(url string) (filePath string, isMatched bool) {
		filenamePrefix := common.Get1stGroupMatch(url, config.ArtworkIDRe)
		if filenamePrefix == "" {
			return filePath, false
		}
		filename := path.Base(url)
		filePath = fmt.Sprintf("%s/%s", config.ThumbnailsFileLocation, filename)
		return filePath, true
	}

	return listenForNetworkEventAndDownloadImages(ctx, urlMatcher)
}

func ListenForNetworkEventAndDownloadArtworkImage(ctx context.Context) (waitFunc func() error) {
	urlMatcher := func(url string) (filePath string, isMatched bool) {
		artworkID := common.Get1stGroupMatch(url, config.ArtworkImgRe)
		if artworkID == "" {
			return filePath, false
		}
		filename := path.Base(url)
		filePath = fmt.Sprintf("%s/%s", config.SavedFileLocation, filename)
		return filePath, true
	}

	return listenForNetworkEventAndDownloadImages(ctx, urlMatcher)
}

func listenForNetworkEventAndDownloadImages(ctx context.Context,
	urlMatcher func(string) (string, bool)) (waitFunc func() error) {
	wg := new(sync.WaitGroup)
	var errs common.Errors
	waitFunc = func() error {
		fmt.Println("waiting writing files to be done")
		wg.Wait()
		return errs.Get()
	}

	eventRespChecker := func(ev *network.EventResponseReceived) bool {
		resp := ev.Response
		if len(resp.Headers) != 0 && string(ev.Type) == `Image` {
			return true
		}
		return false
	}

	//using a mutex to make sure finishing handling one event before the handling next one
	var mutex sync.Mutex
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		//using a go routine to avoid blocking
		go func() {
			mutex.Lock()
			defer mutex.Unlock()
			switch ev := ev.(type) {

			case *network.EventResponseReceived:
				if !eventRespChecker(ev) {
					return
				}

				resp := ev.Response
				filePath, toDownload := urlMatcher(resp.URL)
				if !toDownload {
					return
				}

				wg.Add(1)
				requestID := ev.RequestID
				Manager.RegisterEvent(requestID, func() (selfRemove bool, err error) {
					defer wg.Done()
					err = common.StartSavingResponseToFile(wg, ctx, requestID, filePath)
					return true, err
				})
			case *network.EventLoadingFinished:
				requestID := ev.RequestID
				errs.Add(Manager.TriggerEventIfExist(requestID))
			}
		}()
	})

	return waitFunc
}
