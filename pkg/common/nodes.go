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

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

//TargetNode is used to target a specific node which chromedp does QueryAction on.
//it helps the convoluting usage of "chromedp.FromNode(node.Parent)"
func TargetNode(node *cdp.Node) chromedp.QueryOption {
	return chromedp.ByFunc(func(context.Context, *cdp.Node) ([]cdp.NodeID, error) {
		return []cdp.NodeID{node.NodeID}, nil
	})
}

func GetHref(node *cdp.Node) string {
	attrMap := make(map[string]string)
	attrsLen := len(node.Attributes)
	//odd number index: attr name, even number index: attr value
	for i := 0; i < attrsLen; i += 2 {
		attrMap[node.Attributes[i]] = node.Attributes[i+1]
	}
	return attrMap[config.HrefAttrName]
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
		fmt.Printf("node #%d: nodeType=\"%s\" alt=\"%s\", src=\"%s\"\n", nodeCount, node.NodeType, attrMap[config.AltAttrName], attrMap[config.SrcAttrName])
		nodeCount++
	}
}

func saveScreenshotsOfNodes(ctx context.Context, nodes []*cdp.Node,
	howToSave func([]byte, string) error) (err error) {
	var errs []error
	nodeMap := GetNodesAttrsMap(nodes)
	for _, imgNode := range nodes {
		src := nodeMap[imgNode][config.SrcAttrName]
		filenamePrefix := Get1stGroupMatch(src, config.ArtworkIDRe)
		if filenamePrefix == "" {
			continue
		}
		filename := path.Base(src)
		var buf []byte
		errInner := chromedp.Run(ctx,
			chromedp.Screenshot(config.AnySel, &buf, TargetNode(imgNode)),
		)
		if errInner != nil {
			errs = append(errs, fmt.Errorf("failed to take screenshot of node: %+v", errInner))
			continue
		}
		errInner = howToSave(buf, filename)
		if errInner != nil {
			errs = append(errs, errInner)
		}
	}
	return ConcatenateErrors(errs...)
}

func SaveScreenshotsOfThumbnailNodes(ctx context.Context, imgNodes []*cdp.Node) (err error) {
	howToSave := func(buf []byte, filename string) error {
		filepath := fmt.Sprintf("%s/%s", config.ThumbnailsFileLocation, filename)
		err = ioutil.WriteFile(filepath, buf, config.WriteFilePermission)
		if err != nil {
			fmt.Printf("%s %+v\n", config.ErrorMsgPrefix, err)
			return fmt.Errorf("failed to write to file \"%s\": %+v", filepath, err)
		}
		fmt.Printf("%s wrote %s\n", config.InfMsgPrefix, filepath)
		return nil
	}
	return saveScreenshotsOfNodes(ctx, imgNodes, howToSave)
}

func GetFirstDescendantOfSlibingNodes(node *cdp.Node, descendantNodeLocalName string) *cdp.Node {
	if node.Parent == nil {
		return nil
	}
	if len(node.Parent.Children) <= 0 {
		return nil
	}
	for _, siblingNode := range node.Parent.Children {
		if siblingNode == nil {
			continue
		}
		matchedNode := getFirstDescendantOfNode(siblingNode, descendantNodeLocalName)
		if matchedNode != nil {
			return matchedNode
		}
	}
	return nil
}

func getFirstDescendantOfNode(node *cdp.Node, descendantNodeLocalName string) *cdp.Node {
	for _, childNode := range node.Children {
		if childNode == nil {
			continue
		}
		if childNode.LocalName == descendantNodeLocalName {
			return childNode
		}
		matchedNode := getFirstDescendantOfNode(childNode, descendantNodeLocalName)
		if matchedNode != nil {
			return matchedNode
		}
	}
	return nil
}

func GetNodeWithText(ctx context.Context, textToMatch string, nodes []*cdp.Node) (nodeWithText *cdp.Node, err error) {
	var nodesWithText []*cdp.Node
	for _, node := range nodes {
		if len(node.Children) <= 0 {
			continue
		}
		for _, childNode := range node.Children {
			if childNode == nil {
				continue
			}
			if childNode.NodeType == cdp.NodeTypeText && childNode.NodeValue == textToMatch {
				nodesWithText = append(nodesWithText, node)
				break
			}
		}
	}
	if len(nodesWithText) <= 0 {
		return nodeWithText, fmt.Errorf("no node found with text \"%s\"", textToMatch)
	}
	return nodesWithText[0], nil
}

func getAllNodes(ctx context.Context, sel string,
	selectNode func(*cdp.Node) bool) (nodes []*cdp.Node, err error) {
	var allNodes []*cdp.Node
	err = chromedp.Run(ctx,
		chromedp.Nodes(sel, &allNodes, chromedp.ByQueryAll),
	)
	if err != nil {
		return nodes, fmt.Errorf("failed to get nodes using selector \"%s\": %+v", sel, err)
	}

	if selectNode == nil {
		return allNodes, nil
	}
	for _, node := range allNodes {
		if selectNode(node) {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
}

func GetAllInputNodes(ctx context.Context) (nodes []*cdp.Node, err error) {
	return getAllNodes(ctx, config.InputNodeSel, nil)
}

func GetAllButtonNodes(ctx context.Context) (nodes []*cdp.Node, err error) {
	return getAllNodes(ctx, config.ButtonNodeSel, nil)
}

func GetAllAnchorNodes(ctx context.Context) (nodes []*cdp.Node, err error) {
	return getAllNodes(ctx, config.AnchorNodeSel, nil)
}

func GetAllImgNodes(ctx context.Context) (nodes []*cdp.Node, err error) {
	return getAllNodes(ctx, config.ImgNodeSel, nil)
}

func GetAllSvgNodes(ctx context.Context) (nodes []*cdp.Node, err error) {
	return getAllNodes(ctx, config.SvgNodeSel, nil)
}

func GetAllDivNodes(ctx context.Context) (nodes []*cdp.Node, err error) {
	return getAllNodes(ctx, config.DivNodeSel, nil)
}
