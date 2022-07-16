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
	"github.com/chromedp/cdproto/cdp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

type UrlMap map[string]struct{}

func NewUrlMap() UrlMap {
	return make(map[string]struct{})
}

func (um UrlMap) AddUrlsFromImgNodes(nodes []*cdp.Node) {
	um.AddUrlsFromNodes(nodes, config.SrcAttrName)
}

func (um UrlMap) AddUrlsFromAnchorNodes(nodes []*cdp.Node) {
	um.AddUrlsFromNodes(nodes, config.HrefAttrName)
}

func (um UrlMap) AddUrlsFromNodes(nodes []*cdp.Node, attrName string) {
	for _, node := range nodes {
		um[node.AttributeValue(attrName)] = struct{}{}
	}
}

func (um UrlMap) Aggregate(um2 UrlMap) {
	if um2 == nil {
		return
	}
	for k, v := range um2 {
		um[k] = v
	}
}
