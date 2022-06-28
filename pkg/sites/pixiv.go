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
package sites

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/common"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

var (
	writeFilesWg = new(sync.WaitGroup)
)

func listenForNetworkEventAndDownloadBookmarkThumbnails(ctx context.Context) {
	urlMatcher := func(url string) (filePath string, isMatched bool) {
		filenamePrefix := common.Get1stGroupMatch(url, config.ArtworkIDRe)
		if filenamePrefix == "" {
			return filePath, false
		}
		filename := path.Base(url)
		filePath = fmt.Sprintf("%s/%s", config.SavedFileLocation, filename)
		return filePath, true
	}

	common.ListenForNetworkEventAndDownloadImages(ctx, writeFilesWg, urlMatcher)
}

// func clickOnFigureNodeAndDownloadFullResImgs(ctx context.Context, node *cdp.Node) {

// }

// func clickOnBookmarkImgNodeAndDownloadFullResImgs(ctx context.Context, node *cdp.Node) {
// 	newTabCtx, cancel := chromedp.NewContext(ctx)
// 	defer cancel()
// 	var figureNodes []*cdp.Node
// 	err := chromedp.Run(ctx,
// 		chromedp.MouseClickNode(node),
// 		chromedp.Sleep(2*time.Second),
// 		// get figure nodes
// 		chromedp.Nodes(figureNodeSel, &figureNodes),
// 	)
// 	if len(figureNodes) <= 0 {
// 		return
// 	}
// }

// func clickOnBookmarkImgNodesAndDownloadFullResImgs(ctx context.Context, nodes []*cdp.Node) {

// }

func logoutPixiv(ctx context.Context, screenshotBuf *[]byte) {

	dropdownMenuSel := `#root > div:nth-child(2) > div.sc-12xjnzy-0.dIGjtZ > div:nth-child(1) > div:nth-child(1) > div > div.sc-4nj1pr-3.bWvcqZ > div.sc-4nj1pr-4.jlGtrR > div.sc-pkfh0q-0.kYDpSN > div > button > div`
	logoutButtonSel := `body > div:nth-child(30) > div > div > div > div > ul > li:nth-child(20) > button`
	confirmButtonSel := `body > div:nth-child(30) > div > div > div > div > div > div.sc-hpll47-0.gsvGzp > div.sc-1e6u418-2.fbaJrt > div > div > button.sc-13xx43k-0.sc-13xx43k-1.BSrHG.eGjXJv`

	bookmarkPageUrl := fmt.Sprintf("%s/users/%d/bookmarks/artworks", config.PixivSiteUrl, config.Config.UserID)
	err := chromedp.Run(ctx,
		// go to bookmarks
		chromedp.Navigate(bookmarkPageUrl),
		// to logout
		// open dropdown menu
		chromedp.WaitVisible(dropdownMenuSel),
		chromedp.Click(dropdownMenuSel, chromedp.NodeVisible),
		// click logout
		chromedp.WaitVisible(logoutButtonSel),
		chromedp.Click(logoutButtonSel, chromedp.NodeVisible),
		// click confirm logout
		chromedp.WaitVisible(confirmButtonSel),
		chromedp.Click(confirmButtonSel, chromedp.NodeVisible),

		// just wait
		chromedp.Sleep(2*time.Second),
		// take screenshot
		chromedp.FullScreenshot(screenshotBuf, 90),
	)
	if err != nil {
		log.Fatal(err)
	}
}

func printBookmarkPage(ctx context.Context, bookmarkPage string, screenshotBuf *[]byte) (nextBookmarkPage string) {
	var thumbnailNodes []*cdp.Node
	var nextPageButtons []*cdp.Node

	listenForNetworkEventAndDownloadBookmarkThumbnails(ctx)

	err := chromedp.Run(ctx,
		// go to bookmarks
		chromedp.Navigate(bookmarkPage),
		// just wait
		chromedp.Sleep(5*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	// scroll to the bottom
	err = common.ScrollToButtomOfPage(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = chromedp.Run(ctx,
		// just wait
		chromedp.Sleep(2*time.Second),
		// take screenshot
		chromedp.FullScreenshot(screenshotBuf, 90),
		// get thumbnailNodes
		chromedp.Nodes(config.ThumbnailNodeSel, &thumbnailNodes),
		// get next button node
		chromedp.Nodes(`a.sc-d98f2c-0.sc-xhhh7v-2.cCkJiq.sc-xhhh7v-1-filterProps-Styled-Component.kKBslM`, &nextPageButtons),
	)
	if err != nil {
		log.Fatal(err)
	}

	// common.PrintNodesAltAndSrc(thumbnailNodes)
	// common.SaveThumbnailsImgs(ctx, thumbnailNodes)

	if len(nextPageButtons) > 0 {
		return common.GetHref(nextPageButtons[0])
	}

	return ""
}

func DoPixiv(ctx context.Context) {
	var buf1 []byte
	buf1Filename := `afterPixivLogin.jpeg`
	var buf2 []byte
	buf2Filename := `pixivBookmarkPage.jpeg`
	var buf3 []byte
	buf3Filename := `afterPixivLogout.jpeg`

	err := loginPixiv(ctx, &buf1)
	if err != nil {
		log.Fatal(err)
	}
	// maxNum := 1
	// bookmarkPageUrl := fmt.Sprintf("%s/users/%d/bookmarks/artworks", config.PixivSiteUrl, config.Config.UserID)
	// i := 1
	// for bookmarkPageUrl != "" && i <= maxNum {
	// 	bookmarkPageUrl = printBookmarkPage(ctx, bookmarkPageUrl, &buf2)
	// 	i++
	// }

	err = chromedp.Run(ctx,
		chromedp.Navigate(`https://www.pixiv.net/artworks/92843638`),
		chromedp.Sleep(time.Second*5),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = downloadArtwork(ctx)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("waiting writing files to be done")
	writeFilesWg.Wait()

	logoutPixiv(ctx, &buf3)

	if err := ioutil.WriteFile(buf1Filename, buf1, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s\n", buf1Filename)

	if err := ioutil.WriteFile(buf2Filename, buf2, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s\n", buf2Filename)

	if err := ioutil.WriteFile(buf3Filename, buf3, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s\n", buf3Filename)
}
