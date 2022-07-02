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
	"path"
	"strconv"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/common"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

func listenForNetworkEventAndDownloadBookmarkThumbnails(ctx context.Context) (waitFunc func()) {
	urlMatcher := func(url string) (filePath string, isMatched bool) {
		filenamePrefix := common.Get1stGroupMatch(url, config.ArtworkIDRe)
		if filenamePrefix == "" {
			return filePath, false
		}
		filename := path.Base(url)
		filePath = fmt.Sprintf("%s/%s", config.ThumbnailsFileLocation, filename)
		return filePath, true
	}

	return common.ListenForNetworkEventAndDownloadImages(ctx, urlMatcher)
}

func printBookmarkPage(ctx context.Context, bookmarkPage string, screenshotBuf *[]byte) (err error) {
	var thumbnailNodes []*cdp.Node

	waitDownload := listenForNetworkEventAndDownloadBookmarkThumbnails(ctx)
	defer func() {
		waitDownload()
	}()

	err = chromedp.Run(ctx,
		// go to bookmarks
		chromedp.Navigate(bookmarkPage),
		// just wait
		chromedp.Sleep(5*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to navigate to bookmark page \"%s\": %+v", bookmarkPage, err)
	}

	// scroll to the bottom
	err = common.ScrollToButtomOfPage(ctx)
	if err != nil {
		return fmt.Errorf("failed to scroll to the bottom of page at \"%s\": %+v", bookmarkPage, err)
	}

	err = chromedp.Run(ctx,
		// just wait
		chromedp.Sleep(2*time.Second),
		// take screenshot
		chromedp.FullScreenshot(screenshotBuf, 90),
		// get thumbnailNodes
		chromedp.Nodes(config.ThumbnailNodeSel, &thumbnailNodes, chromedp.ByQueryAll),
	)
	if err != nil {
		return fmt.Errorf("failed to take screenshot or get thumnail nodes: %+v", err)
	}

	common.PrintNodesAltAndSrc(thumbnailNodes)
	common.SaveScreenshotsOfThumbnailNodes(ctx, thumbnailNodes)

	return nil
}

//return values nextPageSvgNode and err can both be nil at the same time
func getAdjacentPageSvgNode(ctx context.Context) (adjacentPageSvgNodes []*cdp.Node, err error) {
	allSvgNodes, err := common.GetAllSvgNodes(ctx)
	if err != nil {
		return adjacentPageSvgNodes, fmt.Errorf("unable to get all svg nodes: %+v", err)
	}
	//get the svg node that has an anchor parent and a nav grandparent
	var matchedSvgNodes []*cdp.Node
	for _, node := range allSvgNodes {
		if node.Parent == nil {
			continue
		}
		if node.Parent.LocalName != config.AnchorNodeSel {
			continue
		}
		if node.Parent.Parent == nil {
			continue
		}
		if node.Parent.Parent.LocalName != config.NavvNodeSel {
			continue
		}
		matchedSvgNodes = append(matchedSvgNodes, node)
	}
	if len(matchedSvgNodes) <= 0 {
		return adjacentPageSvgNodes, fmt.Errorf(`unable to find a svg node with an anchor parent and a nav grandparent`)
	}

	//get the svg node that has an anchor parent (with aria-disabled="false")
	for _, node := range matchedSvgNodes {
		if node.Parent.AttributeValue(config.AriaDisabledAttrName) != "false" {
			continue
		}
		adjacentPageSvgNodes = append(adjacentPageSvgNodes, node)
	}

	return adjacentPageSvgNodes, nil
}

func getPreviousOrNextPageSvgNode(ctx context.Context,
	pageIndexComparer func(currentIdx, IdxOnButton int) bool) (nextPageSvgNode *cdp.Node, err error) {
	adjacentPageSvgNodes, err := getAdjacentPageSvgNode(ctx)
	if err != nil {
		return nextPageSvgNode, fmt.Errorf("failed to get the adjacent page svg node: %+v", err)
	}

	//it should at least has the previous page or the next page button (svg node)
	if len(adjacentPageSvgNodes) <= 0 {
		return nextPageSvgNode, fmt.Errorf("neither of previous page button or next page button is found")
	}

	var urlstr string
	err = chromedp.Run(ctx,
		chromedp.Location(&urlstr),
	)
	if err != nil {
		return nextPageSvgNode, fmt.Errorf("failed to get the url of current page: %+v", err)
	}
	currentPageIdxStr := common.Get3rdGroupMatch(urlstr, config.UserBookmarkPageUrSuffixlRe)
	if currentPageIdxStr == "" { //1st page
		return adjacentPageSvgNodes[0], nil
	}
	currentPageIdx, err := strconv.Atoi(currentPageIdxStr)
	if err != nil {
		return nextPageSvgNode, fmt.Errorf("failed to convert \"%s\" to an integer: %+v", currentPageIdxStr, err)
	}

	var anchorNodes []*cdp.Node
	for _, svgNode := range adjacentPageSvgNodes {
		anchorNodes = append(anchorNodes, svgNode.Parent)
	}
	nodesAttrsMap := common.GetNodesAttrsMap(anchorNodes)
	for node, attrs := range nodesAttrsMap {
		hrefVal := attrs[config.HrefAttrName]
		pageIdxStr := common.Get3rdGroupMatch(hrefVal, config.UserBookmarkPageUrSuffixlRe)
		if pageIdxStr == "" {
			return nextPageSvgNode, fmt.Errorf("invalid %s attr value \"%s\" of parent node of svg node", config.HrefAttrName, hrefVal)
		}
		pageIdx, err := strconv.Atoi(pageIdxStr)
		if err != nil {
			return nextPageSvgNode, fmt.Errorf("failed to convert \"%s\" to an integer: %+v", pageIdxStr, err)
		}
		if pageIndexComparer(currentPageIdx, pageIdx) {
			return node, nil
		}
	}
	return nextPageSvgNode, nil
}

func getPreviousPageSvgNode(ctx context.Context) (nextPageSvgNode *cdp.Node, err error) {
	pageIndexComparer := func(currentIdx, IdxOnButton int) bool {
		return IdxOnButton < currentIdx
	}
	return getPreviousOrNextPageSvgNode(ctx, pageIndexComparer)
}

func getNextPageSvgNode(ctx context.Context) (nextPageSvgNode *cdp.Node, err error) {
	pageIndexComparer := func(currentIdx, IdxOnButton int) bool {
		return IdxOnButton > currentIdx
	}
	return getPreviousOrNextPageSvgNode(ctx, pageIndexComparer)
}

func goToNextBookmarkPage(ctx context.Context) (noNext bool, err error) {
	nextPageSvgNode, err := getNextPageSvgNode(ctx)
	if err != nil {
		return noNext, fmt.Errorf("unable to find the next page button: %+v", err)
	}
	if nextPageSvgNode == nil {
		return true, nil
	}

	err = chromedp.Run(ctx,
		chromedp.MouseClickNode(nextPageSvgNode),
		chromedp.WaitVisible(config.TopLeftPixivImgSel, chromedp.ByQuery),
		//just wait
		chromedp.Sleep(time.Second),
	)

	if err != nil {
		return noNext, fmt.Errorf("failed to click on the next bookmark page button and wait for page to be loaded: %+v", err)
	}
	return noNext, nil
}

func goToNextBookmarkPageAndScrollToTheButtom(ctx context.Context) (noNext bool, err error) {
	noNext, err = goToNextBookmarkPage(ctx)
	if err != nil {
		return noNext, fmt.Errorf("failed to go to next bookmark page: %+v", err)
	}

	if noNext {
		return noNext, nil
	}

	err = common.ScrollToButtomOfPage(ctx)
	if err != nil {
		return noNext, fmt.Errorf("unable to scroll to the buttom of page: %+v", err)
	}

	//just wait for some time for all bookmark items thumbnails to be loaded
	err = chromedp.Run(ctx,
		chromedp.Sleep(time.Second*3),
	)
	return noNext, err
}

func getBookmarkAnchorNode(ctx context.Context) (bookmarkAnchorNode *cdp.Node, err error) {
	nodes, err := common.GetAllAnchorNodes(ctx)
	if err != nil {
		return bookmarkAnchorNode, fmt.Errorf("unable to get all anchor nodes: %+v", err)
	}
	return common.GetNodeWithText(ctx, config.Config.BookmarkAnchorText, nodes)
}

func goToBookmarkPage(ctx context.Context) (err error) {
	err = clickUserProfileImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to click on user profile image: %+v", err)
	}

	bmAnchorNode, err := getBookmarkAnchorNode(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bookmark anchor node: %+v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.MouseClickNode(bmAnchorNode),
		chromedp.WaitVisible(config.TopLeftPixivImgSel, chromedp.ByQuery),
		// just wait
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to click on to bookmark anchor: %+v", err)
	}
	return nil
}

func goToBookmarkPageAndScrollToTheButtom(ctx context.Context) (err error) {
	err = goToBookmarkPage(ctx)
	if err != nil {
		return fmt.Errorf("failed to go to bookmark page: %+v", err)
	}

	// dismiss tutorials banners
	err = dismissTutorialBanner(ctx)
	if err != nil {
		return fmt.Errorf("failed to dismiss tutorial banners: %+v", err)
	}

	err = common.ScrollToButtomOfPage(ctx)
	if err != nil {
		return fmt.Errorf("unable to scroll to the buttom of page: %+v", err)
	}

	//just wait for some time for all bookmark items thumbnails to be loaded
	return chromedp.Run(ctx,
		chromedp.Sleep(time.Second*3),
	)
}

func getUserID(ctx context.Context) (err error) {
	var urlstr string
	err = chromedp.Run(ctx,
		chromedp.Location(&urlstr),
	)
	if err != nil {
		return fmt.Errorf("failed to get the url of current page: %+v", err)
	}
	userID := common.Get1stGroupMatch(urlstr, config.UserBookmarkPageUrSuffixlRe)
	if userID == "" {
		return fmt.Errorf("no 1st group match from \"%s\" using regex \"%s\"", urlstr, config.UserBookmarkPageUrSuffixlRe.String())
	}
	config.Config.UserID = userID
	return nil
}

func iterateBookmarkPages(ctx context.Context, maxIteration int,
	toDo func(context.Context) error) (err error) {

	waitDownload := listenForNetworkEventAndDownloadBookmarkThumbnails(ctx)
	defer func() {
		waitDownload()
	}()

	err = goToBookmarkPageAndScrollToTheButtom(ctx)
	if err != nil {
		return fmt.Errorf("failed to go to bookmark page and scroll to the bottom: %+v", err)
	}

	warning := getUserID(ctx)
	if warning != nil {
		fmt.Printf("failed to get user ID: %+v", warning)
	}

	if toDo != nil {
		err = toDo(ctx)
		if err != nil {
			return fmt.Errorf("failed to do toDo(): %+v", err)
		}
	}

	ithIteration := 2 //you will be on the 2nd page the 1st time when you click the next page button
	toContinue := func() bool {
		if maxIteration <= 0 {
			return true
		}
		if ithIteration < maxIteration {
			return true
		}
		return false
	}

	var noNext bool
	for !noNext && toContinue() {
		noNext, err = goToNextBookmarkPageAndScrollToTheButtom(ctx)
		if err != nil {
			return fmt.Errorf("failed to go to next bookmark page and scroll to the bottom: %+v", err)
		}
		if toDo != nil {
			err = toDo(ctx)
			if err != nil {
				return fmt.Errorf("failed to do toDo(): %+v", err)
			}
		}
		ithIteration++
	}
	return nil
}

func getBookmarkItemThumbnailNodes(ctx context.Context) (imgNodes []*cdp.Node, err error) {
	nodes, err := common.GetAllImgNodes(ctx)
	if err != nil {
		return imgNodes, fmt.Errorf("failed to get all img nodes: %+v", err)
	}
	nodesAttrsMap := common.GetNodesAttrsMap(nodes)
	for node, attrs := range nodesAttrsMap {
		if node.Parent == nil {
			continue
		}
		if node.Parent.LocalName != config.DivNodeSel {
			continue
		}
		if node.Parent.Parent == nil {
			continue
		}
		if node.Parent.Parent.LocalName != config.AnchorNodeSel {
			continue
		}
		srcVal := attrs[config.SrcAttrName]
		filenamePrefix := common.Get1stGroupMatch(srcVal, config.ArtworkIDRe)
		if filenamePrefix == "" {
			continue
		}
		hrefVal := node.Parent.Parent.AttributeValue(config.HrefAttrName)
		artworkID := common.Get1stGroupMatch(hrefVal, config.ArkworkerUrlSuffixRe)
		if artworkID == "" {
			continue
		}
		imgNodes = append(imgNodes, node)
	}
	return imgNodes, nil
}

func getBookmarkItemAnchorNodes(ctx context.Context) (anchorNodes []*cdp.Node, err error) {
	imgNodes, err := getBookmarkItemThumbnailNodes(ctx)
	if err != nil {
		return anchorNodes, fmt.Errorf("failed to get thumbnail img nodes: %v", err)
	}
	for _, imgNode := range imgNodes {
		anchorNodes = append(anchorNodes, imgNode.Parent.Parent)
	}
	return anchorNodes, nil
}

func openBookmarkItemInNewTab(ctx context.Context,
	toDo func(context.Context) error) (err error) {
	anchorNodes, err := getBookmarkItemAnchorNodes(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bookmark item anchor nodes: %+v", err)
	}
	for _, node := range anchorNodes {
		err = func() error {
			newTabCtx, cancel, err := common.ClickOnAnchorAndOpenNewTab(ctx, node)
			defer func() {
				cancel()
			}()
			if err != nil {
				return fmt.Errorf("failed to click on anchor and open new tab: %+v", err)
			}
			//wait for some time for the page to be loaded
			chromedp.Run(newTabCtx,
				chromedp.Sleep(2*time.Second),
			)
			if toDo != nil {
				err = toDo(newTabCtx)
				if err != nil {
					return fmt.Errorf("failed to do toDo(): %+v", err)
				}
			}
			return nil
		}()
		if err != nil {
			return err
		}
	}
	return nil
}

func getCloseTutorialBannerButton(ctx context.Context) (closeButton *cdp.Node, err error) {
	maxRetry := 2
	i := 0
	for closeButton == nil && i < maxRetry {
		nodes, err := common.GetAllDivNodes(ctx)
		if err != nil {
			return closeButton, fmt.Errorf("failed to get all div nodes: %+v", err)
		}
		bannerTextDivNode, err := common.GetNodeWithText(ctx, config.Config.BookmarkTutorialBannerText, nodes)
		if err != nil {
			return closeButton, fmt.Errorf("failed to get node with text \"%s\": %+v", config.Config.BookmarkTutorialBannerText, err)
		}

		err = common.RequestSubtree(ctx, bannerTextDivNode.Parent)
		if err != nil {
			return closeButton, fmt.Errorf("failed to request subtree of tutorial banner: %+v", err)
		}

		closeButton = common.GetFirstDescendantOfSlibingNodes(bannerTextDivNode, config.SvgNodeSel)
		if closeButton != nil {
			return closeButton, nil
		}
		i++
	}
	return closeButton, fmt.Errorf("no close svg node found after %d retry", maxRetry)
}

func dismissTutorialBanner(ctx context.Context) (err error) {
	closeButton, err := getCloseTutorialBannerButton(ctx)
	if err != nil {
		return fmt.Errorf("failed get the close button of tutorial banner: %+v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.MouseClickNode(closeButton),
		chromedp.Sleep(time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to close button of tutorial banner: %+v", err)
	}
	return nil
}
