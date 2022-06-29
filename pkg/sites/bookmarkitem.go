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
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/common"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

const (
	artworkerImageAnchorRelVal = `noopener`
)

func getAnchorNodeOfArtworkImg(ctx context.Context) (anchor *cdp.Node, multiImgs bool, err error) {
	anchorNodes, err := common.GetAllAnchorNodes(ctx)
	if err != nil {
		return anchor, multiImgs, fmt.Errorf("failed to get all anchor nodes: %+v", err)
	}

	nodesAttrsMap := common.GetNodesAttrsMap(anchorNodes)
	for node, attrs := range nodesAttrsMap {
		relVal := attrs[config.RelAttrName]
		if relVal != artworkerImageAnchorRelVal {
			continue
		}
		hrefVal := attrs[config.HrefAttrName]
		artworkID := common.Get1stGroupMatch(hrefVal, config.ArtworkImgRe)
		if artworkID == "" {
			continue
		}
		classVal := attrs[config.ClassAttrName]
		if !strings.Contains(classVal, `gtm-expand-full-size-illust`) {
			multiImgs = true
		}
		return node, multiImgs, nil
	}

	return anchor, multiImgs, fmt.Errorf("no node with attr \"%s\" value equal to \"%s\" and has href to an artwork", config.RelAttrName, artworkerImageAnchorRelVal)
}

func getAnchorNodesOfArtworkImg(ctx context.Context) (anchors []*cdp.Node, err error) {
	anchorNodes, err := common.GetAllAnchorNodes(ctx)
	if err != nil {
		return anchors, fmt.Errorf("failed to get all anchor nodes: %+v", err)
	}

	nodesAttrsMap := common.GetNodesAttrsMap(anchorNodes)
	for node, attrs := range nodesAttrsMap {
		relVal := attrs[config.RelAttrName]
		if relVal != artworkerImageAnchorRelVal {
			continue
		}
		hrefVal := attrs[config.HrefAttrName]
		artworkID := common.Get1stGroupMatch(hrefVal, config.ArtworkImgRe)
		if artworkID == "" {
			continue
		}
		classVal := attrs[config.ClassAttrName]
		if strings.Contains(classVal, `gtm-expand-full-size-illust`) {
			anchors = append(anchors, node)
		}
	}
	return anchors, nil
}

func getArtworkImgNode(ctx context.Context) (imgNode *cdp.Node, err error) {
	imgNodes, err := common.GetAllImgNodes(ctx)
	if err != nil {
		return imgNode, fmt.Errorf("failed to get all img nodes: %+v", err)
	}

	nodesAttrsMap := common.GetNodesAttrsMap(imgNodes)
	for node, attrs := range nodesAttrsMap {
		srcVal := attrs[config.SrcAttrName]
		artworkID := common.Get1stGroupMatch(srcVal, config.ArtworkImgRe)
		if artworkID == "" {
			continue
		}
		return node, nil
	}

	return imgNode, fmt.Errorf("no img node has \"%s\" attr value matchs regex \"%s\"", config.SrcAttrName, config.ArtworkImgRe.String())
}

func listenForNetworkEventAndDownloadArtworkImage(ctx context.Context, wg *sync.WaitGroup) {
	urlMatcher := func(url string) (filePath string, isMatched bool) {
		artworkID := common.Get1stGroupMatch(url, config.ArtworkImgRe)
		if artworkID == "" {
			return filePath, false
		}
		filename := path.Base(url)
		filePath = fmt.Sprintf("%s/%s", config.SavedFileLocation, filename)
		return filePath, true
	}

	common.ListenForNetworkEventAndDownloadImages(ctx, wg, urlMatcher)
}

func downloadMultiImgsArtwork(ctx context.Context, anchorNode *cdp.Node) (err error) {
	err = chromedp.Run(ctx,
		chromedp.Click(config.ImgNodeSel, chromedp.ByQuery, chromedp.FromNode(anchorNode)),
		chromedp.Sleep(config.SavingRespWaitDura),
	)
	if err != nil {
		return fmt.Errorf("failed to click on artwork image: %+v", err)
	}

	anchorNodes, err := getAnchorNodesOfArtworkImg(ctx)
	if err != nil {
		return fmt.Errorf("failed to get the list of anchor nodes: %+v", err)
	}
	fmt.Printf("number of anchorNodes: %d\n", len(anchorNodes))
	for _, anchorNode := range anchorNodes {

		var imgNode *cdp.Node
		for imgNode == nil {
			//keep clicking until it actually zoomed into the full res image for working wround two different clicking behaviors (move up/down or zoom in)
			//to zoom in
			err := chromedp.Run(ctx,
				chromedp.MouseClickNode(anchorNode),
				chromedp.Sleep(config.SavingRespWaitDura),
			)
			if err != nil {
				return fmt.Errorf("failed to click on artwork image: %+v", err)
			}
			//to zoom out
			imgNode, err = getArtworkImgNode(ctx)
			if err != nil {
				fmt.Printf("warning: unable to find img node for full res artwork: %+v. retrying\n", err)
			}
		}

		err = chromedp.Run(ctx,
			chromedp.Click(config.ImgNodeSel, chromedp.ByQuery, chromedp.FromNode(imgNode.Parent)),
			chromedp.Sleep(time.Second),
		)
		if err != nil {
			return fmt.Errorf("unable to click on img node for artwork: %+v", err)
		}
	}
	return nil
}

func downloadSingleImgArtwork(ctx context.Context, anchorNode *cdp.Node) (err error) {
	return chromedp.Run(ctx,
		chromedp.Click(config.ImgNodeSel, chromedp.ByQuery, chromedp.FromNode(anchorNode)),
		chromedp.Sleep(config.SavingRespWaitDura),
	)

}

func downloadArtwork(ctx context.Context, wg *sync.WaitGroup) (err error) {
	anchorNode, multiImgs, err := getAnchorNodeOfArtworkImg(ctx)
	if err != nil {
		return fmt.Errorf("unable to find anchor node of artwork: %+v", err)
	}
	listenForNetworkEventAndDownloadArtworkImage(ctx, wg)
	if multiImgs {
		err = downloadMultiImgsArtwork(ctx, anchorNode)
	} else {
		err = downloadSingleImgArtwork(ctx, anchorNode)
	}
	if err != nil {
		return fmt.Errorf("failed to click on artwork image: %+v", err)
	}
	return nil
}

func navigateToArtworkPageAndDownloadArtwork(ctx context.Context, url string) (err error) {
	err = chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(time.Second*5),
	)
	if err != nil {
		return fmt.Errorf("failed to navigate to \"%s\": %+v", url, err)
	}

	wg := new(sync.WaitGroup)
	defer func() {
		fmt.Println("waiting writing files to be done")
		wg.Wait()
	}()
	err = downloadArtwork(ctx, wg)
	if err != nil {
		return fmt.Errorf("failed to download artwork at \"%s\"", url)
	}

	return nil
}
