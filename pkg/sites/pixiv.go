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
	"log"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/common"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

func getSubmitButtonNode(ctx context.Context, buttonText string) (submitButtonNode *cdp.Node, err error) {
	nodes, err := common.GetAllButtonNodes(ctx)
	if err != nil {
		return submitButtonNode, fmt.Errorf("unable to get all button nodes: %+v", err)
	}
	nodesAttrsMap := common.GetNodesAttrsMap(nodes)
	var submitButtonNodes []*cdp.Node
	for node, attrs := range nodesAttrsMap {
		typeVal := attrs[config.TypeAttrName]
		if typeVal == submitButtonTypeAttrVal {
			submitButtonNodes = append(submitButtonNodes, node)
		}
	}
	if len(submitButtonNodes) <= 0 {
		return submitButtonNode, fmt.Errorf("no button node found with %s attr value = \"%s\"", config.TypeAttrName, submitButtonTypeAttrVal)
	}

	return getNodeWithText(ctx, buttonText, submitButtonNodes)
}

func DoPixiv(ctx context.Context) {
	err := loginPixiv(ctx)
	if err != nil {
		log.Fatal(err)
	}

	err = iterateBookmarkPages(ctx, 10)
	if err != nil {
		log.Println(err)
	}

	// err = navigateToArtworkPageAndDownloadArtwork(ctx, `https://www.pixiv.net/artworks/92843638`)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	err = logoutPixiv(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func getNodeWithText(ctx context.Context, textToMatch string, nodes []*cdp.Node) (nodeWithText *cdp.Node, err error) {
	for _, node := range nodes {
		var text string
		err = chromedp.Run(ctx,
			chromedp.Text(config.AnySel, &text, chromedp.ByQuery, common.TargetNode(node)),
		)
		if err != nil {
			return nodeWithText, fmt.Errorf("failed to get text from node: %+v", err)
		}
		if text == textToMatch {
			return node, nil
		}
	}
	return nodeWithText, fmt.Errorf("no node found with text \"%s\"", textToMatch)
}

func getUserProfileImgNode(ctx context.Context) (userProfileImgNode *cdp.Node, err error) {
	nodes, err := common.GetAllImgNodes(ctx)
	if err != nil {
		return userProfileImgNode, fmt.Errorf("unable to get all img nodes: %+v", err)
	}
	nodesAttrsMap := common.GetNodesAttrsMap(nodes)
	//img node should have a div parent and a button grandparrent as well as src value matchs profile picture regex
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
		if node.Parent.Parent.LocalName != config.ButtonNodeSel {
			continue
		}
		srcVal := attrs[config.SrcAttrName]
		srcSuffix := common.Get1stGroupMatch(srcVal, config.UserProfileImgSrcRe)
		if srcSuffix == "" {
			continue
		}
		return node, nil
	}
	return userProfileImgNode, fmt.Errorf("no img node has \"%s\" attr value matchs regex \"%s\"", config.SrcAttrName, config.UserProfileImgSrcRe.String())
}

func clickUserProfileImage(ctx context.Context) (err error) {
	profileImgNode, err := getUserProfileImgNode(ctx)
	if err != nil {
		return fmt.Errorf("failed to get profile img node: %+v", err)
	}

	return chromedp.Run(ctx,
		chromedp.MouseClickNode(profileImgNode),
		// just wait
		chromedp.Sleep(1*time.Second),
	)
}
