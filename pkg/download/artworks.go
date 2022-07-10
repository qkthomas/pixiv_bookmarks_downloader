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

func ListenForNetworkEventAndDownloadBookmarkThumbnails(ctx context.Context) (waitFunc func(int) error) {
	//do not do duplicate download
	var mutex sync.Mutex
	downloadedUrls := make(map[string]struct{})
	checkDuplicate := func(url string) (toDownload bool) {
		mutex.Lock()
		defer mutex.Unlock()
		_, wasDownloaded := downloadedUrls[url]
		if wasDownloaded {
			return false
		}
		downloadedUrls[url] = struct{}{}
		return true
	}
	urlMatcher := func(url string) (filePath string, isMatched bool) {
		toDownload := checkDuplicate(url)
		if !toDownload {
			return filePath, false
		}
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

func ListenForNetworkEventAndDownloadArtworkImage(ctx context.Context) (waitFunc func(int) error) {
	//do not do duplicate download
	var mutex sync.Mutex
	downloadedUrls := make(map[string]struct{})
	checkDuplicate := func(url string) (toDownload bool) {
		mutex.Lock()
		defer mutex.Unlock()
		_, wasDownloaded := downloadedUrls[url]
		if wasDownloaded {
			return false
		}
		downloadedUrls[url] = struct{}{}
		return true
	}
	urlMatcher := func(url string) (filePath string, isMatched bool) {
		toDownload := checkDuplicate(url)
		if !toDownload {
			return filePath, false
		}
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
	urlMatcher func(string) (string, bool)) (waitFunc func(int) error) {

	waitItemChan := make(chan struct{}, 100)
	var errs common.Errors
	waitFunc = func(numberOfImg int) error {
		fmt.Printf("waiting writing %d files to be done\n", numberOfImg)
		fmt.Printf("debug: len(waitItemChan)=%d\n", len(waitItemChan))
		for i := 1; i <= numberOfImg; i++ {
			_ = <-waitItemChan
		}
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

				requestID := ev.RequestID
				fmt.Printf("registering event: requestID: \"%s\", url=\"%s\"\n", requestID, resp.URL)
				Manager.RegisterEvent(requestID, func() (selfRemove bool, err error) {
					defer func() {
						waitItemChan <- struct{}{}
					}()
					fmt.Printf("start writing to file: requestID: \"%s\", filePath=\"%s\"\n", requestID, filePath)
					err = common.StartSavingResponseToFile(ctx, requestID, filePath)
					fmt.Printf("finish writing to file: requestID: \"%s\", filePath=\"%s\"\n", requestID, filePath)
					return true, err
				})
			case *network.EventLoadingFinished:
				requestID := ev.RequestID
				fmt.Printf("trigger event: requestID: \"%s\"\n", requestID)
				errs.Add(Manager.TriggerEventIfExist(requestID))
			}
		}()
	})

	return waitFunc
}
