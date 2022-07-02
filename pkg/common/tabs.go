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

/*
	using examples from: https://github.com/chromedp/chromedp/issues/835
*/
package common

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

func ClickOnAnchorAndOpenNewTab(ctx context.Context, node *cdp.Node) (newTabCtx context.Context, cancelFunc func(), err error) {
	cancelFunc = func() {} //panic guard
	if node.LocalName != config.AnchorNodeSel {
		return newTabCtx, cancelFunc, fmt.Errorf("the node is a \"%s\", not an anchor", node.LocalName)
	}

	hrefVal := node.AttributeValue(config.HrefAttrName)
	if hrefVal == "" {
		return newTabCtx, cancelFunc, fmt.Errorf("the anchor node has not %s attr value", config.HrefAttrName)
	}

	err = chromedp.Run(ctx,
		chromedp.MouseClickNode(node, chromedp.ButtonModifiers(config.NewTabMouseClickModifier)),
	)
	if err != nil {
		return newTabCtx, cancelFunc, fmt.Errorf("unable to click on node with mouse modifier \"%s\": %+v", config.NewTabMouseClickModifier.String(), err)
	}

	targets, err := chromedp.Targets(ctx)
	if err != nil {
		return newTabCtx, cancelFunc, fmt.Errorf("failed to get all targets from current context: %+v", err)
	}

	for _, t := range targets {
		if strings.HasSuffix(t.URL, hrefVal) {
			newTabCtx, cancelFunc = chromedp.NewContext(ctx, chromedp.WithTargetID(t.TargetID))
			// chromedp.Run is required so that it will send Target.attachToTarget command
			err = chromedp.Run(newTabCtx)
			if err != nil {
				return newTabCtx, cancelFunc, fmt.Errorf("failed to init new tab context: %+v", err)
			}
			c := chromedp.FromContext(ctx) //both ctx and newTabCtx work
			err = target.ActivateTarget(t.TargetID).Do(cdp.WithExecutor(ctx, c.Target))
			if err != nil {
				return newTabCtx, cancelFunc, fmt.Errorf("failed to active target (id %s): %+v", t.TargetID, err)
			}
			sleepTime := time.Second
			err = chromedp.Run(newTabCtx,
				chromedp.Sleep(sleepTime),
			)
			if err != nil {
				return newTabCtx, cancelFunc, fmt.Errorf("unable to sleep for \"%s\": %+v", sleepTime.String(), err)
			}
			return newTabCtx, cancelFunc, nil
		}
	}
	return newTabCtx, cancelFunc, fmt.Errorf("unable to find target has url suffix \"%s\"", hrefVal)
}
