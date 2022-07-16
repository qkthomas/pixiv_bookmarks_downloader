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

func ListenForNetworkEventAndDownloadBookmarkThumbnails(ctx context.Context) (waitFunc func(common.UrlMap) error) {
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

func ListenForNetworkEventAndDownloadArtworkImage(ctx context.Context) (waitFunc func(common.UrlMap) error) {
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
	urlMatcher func(string) (string, bool)) (waitFunc func(common.UrlMap) error) {

	waitItemChan := make(chan string, 100)
	var errs common.Errors
	eventQueue := make(chan interface{}, 1000) //how big is enough?
	var mutex sync.Mutex
	var isEventQueueClosed bool
	waitFunc = func(urls common.UrlMap) error {
		defer func() {
			mutex.Lock()
			close(eventQueue) //will it cause panic?
			isEventQueueClosed = true
			mutex.Unlock()
		}()
		fmt.Printf("waiting writing %d files to be done\n", len(urls))
		fmt.Printf("debug: len(waitItemChan)=%d\n", len(waitItemChan))
		for len(urls) > 0 {
			url := <-waitItemChan
			_, ok := urls[url]
			if ok {
				delete(urls, url)
			}
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

	go func() {
		for ev := range eventQueue {
			switch ev := ev.(type) {
			case *network.EventResponseReceived:
				if !eventRespChecker(ev) {
					continue
				}

				resp := ev.Response
				url := resp.URL
				filePath, toDownload := urlMatcher(url)
				if !toDownload {
					continue
				}

				requestID := ev.RequestID
				fmt.Printf("registering event: requestID: \"%s\", url=\"%s\"\n", requestID, url)
				Manager.RegisterEvent(requestID, func() (selfRemove bool, err error) {
					defer func() {
						waitItemChan <- url
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
		}
	}()

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		mutex.Lock()
		defer mutex.Unlock()
		if isEventQueueClosed {
			return
		}
		switch ev := ev.(type) {
		case *network.EventResponseReceived:
			eventQueue <- ev
		case *network.EventLoadingFinished:
			eventQueue <- ev
		}
	})

	return waitFunc
}
