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
	"sync"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

func StartSavingResponseToFile(wg *sync.WaitGroup, ctx context.Context, requestID network.RequestID, filepath string) (err error) {
	param := network.GetResponseBody(requestID)
	if param == nil {
		return
	}
	c := chromedp.FromContext(ctx)
	buf, err := param.Do(cdp.WithExecutor(ctx, c.Target))
	if err != nil {
		return fmt.Errorf("error when doing param.Do(ctx): %+v", err)
	}
	fmt.Printf("writing to file \"%s\"\n", filepath)
	if err = ioutil.WriteFile(filepath, buf, config.WriteFilePermission); err != nil {
		return fmt.Errorf("error: failed to write to %s: %+v", filepath, err)
	}
	fmt.Printf("wrote %s\n", filepath)
	return nil
}

func LogPageLoaded(ctx context.Context) {
	c := chromedp.FromContext(ctx)
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		case *page.EventLoadEventFired:
			fmt.Printf("EventLoadEventFired for target ID \"%s\" at %s\n", c.Target.TargetID, ev.Timestamp.Time().String())
			// other needed network Event
		}
	})
}
