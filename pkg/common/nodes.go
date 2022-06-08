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
	"path"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

func GetHref(node *cdp.Node) string {
	attrMap := make(map[string]string)
	attrsLen := len(node.Attributes)
	//odd number index: attr name, even number index: attr value
	for i := 0; i < attrsLen; i += 2 {
		attrMap[node.Attributes[i]] = node.Attributes[i+1]
	}
	return attrMap["href"]
}

func GetNodesAttrsMap(nodes []*cdp.Node) map[*cdp.Node]map[string]string {
	nodeMap := make(map[*cdp.Node]map[string]string)
	for _, node := range nodes {
		attrMap := make(map[string]string)
		attrsLen := len(node.Attributes)
		//odd number index: attr name, even number index: attr value
		for i := 0; i < attrsLen; i += 2 {
			attrMap[node.Attributes[i]] = node.Attributes[i+1]
		}
		nodeMap[node] = attrMap
	}
	return nodeMap
}

func PrintNodesAltAndSrc(nodes []*cdp.Node) {
	nodeMap := GetNodesAttrsMap(nodes)
	nodeCount := 1
	for node, attrMap := range nodeMap {
		fmt.Printf("node #%d: nodeType=\"%s\" alt=\"%s\", src=\"%s\"\n", nodeCount, node.NodeType, attrMap[`alt`], attrMap[`src`])
		nodeCount++
	}
}

func saveScreenshotsOfNodes(ctx context.Context, nodes []*cdp.Node,
	howToSave func([]byte, string)) {
	nodeMap := GetNodesAttrsMap(nodes)
	for _, imgNode := range nodes {
		parentNode := imgNode.Parent
		src := nodeMap[imgNode][`src`]
		filenamePrefix := Get1stGroupMatch(src, config.ArtworkIDRe)
		if filenamePrefix == "" {
			continue
		}
		filename := path.Base(src)
		var buf []byte
		chromedp.Run(ctx,
			chromedp.Screenshot(config.ThumbnailNodeSel, &buf, chromedp.FromNode(parentNode)),
		)
		howToSave(buf, filename)
	}
}

func SaveThumbnailsImgsToFile(ctx context.Context, imgNodes []*cdp.Node) {
	howToSave := func(buf []byte, filename string) {
		filepath := fmt.Sprintf("%s/%s", config.ThumbnailsFileLocation, filename)
		if err := ioutil.WriteFile(filepath, buf, config.WriteFilePermission); err != nil {
			fmt.Printf("%s %+v\n", config.ErrorMsgPrefix, err)
			return
		}
		fmt.Printf("%s wrote %s\n", config.InfMsgPrefix, filepath)
	}
	saveScreenshotsOfNodes(ctx, imgNodes, howToSave)
}

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

func getAllNodes(ctx context.Context, sel string) (nodes []*cdp.Node, err error) {
	err = chromedp.Run(ctx,
		chromedp.Nodes(sel, &nodes),
	)
	return nodes, err
}

func GetAllInputNodes(ctx context.Context) (nodes []*cdp.Node, err error) {
	return getAllNodes(ctx, config.InputNodeSel)
}

func GetAllButtonNodes(ctx context.Context) (nodes []*cdp.Node, err error) {
	return getAllNodes(ctx, config.ButtonNodeSel)
}
