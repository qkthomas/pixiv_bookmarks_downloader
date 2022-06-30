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
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/common"
	"github.com/qkthomas/pixiv_bookmarks_downloader/pkg/config"
)

func getLogoutButtonNode(ctx context.Context) (logoutButtonNode *cdp.Node, err error) {
	nodes, err := common.GetAllButtonNodes(ctx)
	if err != nil {
		return logoutButtonNode, fmt.Errorf("unable to get all button nodes: %+v", err)
	}
	return getNodeWithText(ctx, config.Config.LogoutButtonText, nodes)
}

func getLogoutConfirmationButtonNode(ctx context.Context) (logoutConfirmationButtonNode *cdp.Node, err error) {
	return getSubmitButtonNode(ctx, config.Config.ConfirmLogoutButtonText)
}

func logoutPixiv(ctx context.Context) (err error) {
	err = clickUserProfileImage(ctx)
	if err != nil {
		return fmt.Errorf("failed to open user profile dropdown menu: %+v", err)
	}

	logoutButton, err := getLogoutButtonNode(ctx)
	if err != nil {
		return fmt.Errorf("unable to find logout button: %+v", err)
	}
	err = chromedp.Run(ctx,
		chromedp.MouseClickNode(logoutButton),
		// just wait
		chromedp.Sleep(1*time.Second),
	)
	if err != nil {
		return fmt.Errorf("unable to click the logout button: %+v", err)
	}

	logoutConfirmationButton, err := getLogoutConfirmationButtonNode(ctx)
	if err != nil {
		return fmt.Errorf("unable to find logout confirmation button: %+v", err)
	}
	err = chromedp.Run(ctx,
		chromedp.MouseClickNode(logoutConfirmationButton),
		// just wait
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		return fmt.Errorf("unable to click the logout button: %+v", err)
	}
	return nil
}
