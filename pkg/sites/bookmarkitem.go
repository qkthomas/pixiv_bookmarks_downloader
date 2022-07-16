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
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/common"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/download"
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
	fmt.Printf("debug: enter getArtworkImgNode()\n")
	defer fmt.Printf("debug: exit getArtworkImgNode()\n")
	imgNodes, err := common.GetAllImgNodes(ctx)
	if err != nil {
		return imgNode, fmt.Errorf("failed to get all img nodes: %+v", err)
	}

	var matchedNodes []*cdp.Node
	nodesAttrsMap := common.GetNodesAttrsMap(imgNodes)
	for node, attrs := range nodesAttrsMap {
		srcVal := attrs[config.SrcAttrName]
		artworkID := common.Get1stGroupMatch(srcVal, config.ArtworkImgRe)
		if artworkID == "" {
			continue
		}
		matchedNodes = append(matchedNodes, node)
	}

	if len(matchedNodes) <= 0 {
		return imgNode, fmt.Errorf("no img node has \"%s\" attr value matchs regex \"%s\"", config.SrcAttrName, config.ArtworkImgRe.String())
	}
	for i, node := range matchedNodes {
		fmt.Printf("debug: matchedNodes[%d].src=\"%s\"\n", i, node.AttributeValue(config.SrcAttrName))
	}
	return matchedNodes[0], nil
}

func downloadMultiImgsArtwork(ctx context.Context, anchorNode *cdp.Node) (urls common.UrlMap, err error) {
	err = chromedp.Run(ctx,
		chromedp.Click(config.ImgNodeSel, chromedp.ByQuery, chromedp.FromNode(anchorNode)),
		chromedp.Sleep(config.SavingRespWaitDura),
	)
	if err != nil {
		return urls, fmt.Errorf("failed to click on artwork image: %+v", err)
	}

	anchorNodes, err := getAnchorNodesOfArtworkImg(ctx)
	if err != nil {
		return urls, fmt.Errorf("failed to get the list of anchor nodes: %+v", err)
	}
	urls = common.NewUrlMap()
	urls.AddUrlsFromAnchorNodes(anchorNodes)

	var errs []error
	for _, anchorNode := range anchorNodes {
		_, er := downloadSingleImgArtwork(ctx, anchorNode)
		errs = append(errs, er)
	}
	return urls, common.ConcatenateErrors(errs...)
}

func downloadSingleImgArtwork(ctx context.Context, anchorNode *cdp.Node) (urls common.UrlMap, err error) {

	urls = common.NewUrlMap()
	urls.AddUrlsFromAnchorNodes([]*cdp.Node{anchorNode})

	var imgNode *cdp.Node
	for {
		//keep clicking until it actually zoomed into the full res image for working wround two different clicking behaviors (move up/down or zoom in)
		//to zoom in
		err = chromedp.Run(ctx,
			chromedp.MouseClickNode(anchorNode),
			chromedp.Sleep(time.Second),
		)
		if err != nil {
			return urls, fmt.Errorf("failed to click on artwork image: %+v", err)
		}
		//double check if the src of img node match the href of the anchor node
		imgNode, err = getArtworkImgNode(ctx)
		if err != nil {
			fmt.Printf("warning: unable to find img node for full res artwork: %+v. retrying\n", err)
			continue
		}

		imgSrc := imgNode.AttributeValue(config.SrcAttrName)
		aHref := anchorNode.AttributeValue(config.HrefAttrName)
		if imgSrc == aHref {
			fmt.Printf("img.src=\"%s\", a.href=\"%s\"\n", imgSrc, aHref)
			break
		}
	}

	err = EscapeFromFullResImg2(ctx)
	if err != nil {
		return urls, fmt.Errorf("unable to click on img node for artwork: %+v", err)
	}

	return urls, nil

}

func downloadArtwork(ctx context.Context) (err error) {
	anchorNode, multiImgs, err := getAnchorNodeOfArtworkImg(ctx)
	if err != nil {
		return fmt.Errorf("unable to find anchor node of artwork: %+v", err)
	}

	var urls common.UrlMap
	waitDownload := download.ListenForNetworkEventAndDownloadArtworkImage(ctx)
	defer func() {
		errs := []error{err, waitDownload(urls)}
		err = common.ConcatenateErrors(errs...)
	}()

	if multiImgs {
		urls, err = downloadMultiImgsArtwork(ctx, anchorNode)
	} else {
		urls, err = downloadSingleImgArtwork(ctx, anchorNode)
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

	err = downloadArtwork(ctx)
	if err != nil {
		return fmt.Errorf("failed to download artwork at \"%s\"", url)
	}

	return nil
}

func EscapeFromFullResImg1(ctx context.Context, imgNode *cdp.Node) (err error) {
	return chromedp.Run(ctx,
		chromedp.MouseClickNode(imgNode),
		chromedp.Sleep(time.Second),
	)
}

func EscapeFromFullResImg2(ctx context.Context) (err error) {
	return chromedp.Run(ctx,
		chromedp.KeyEvent(kb.Escape),
		chromedp.Sleep(time.Second),
	)
}
