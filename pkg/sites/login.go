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

const (
	loginButtonTypeAttrVal = `submit`
)

func getUserAndPasswordInputNodes(ctx context.Context) (userNode *cdp.Node, passwordNode *cdp.Node, err error) {
	nodes, err := common.GetAllInputNodes(ctx)
	if err != nil {
		return userNode, passwordNode, fmt.Errorf("unable to get all input nodes: %+v", err)
	}

	nodesAttrsMap := common.GetNodesAttrsMap(nodes)
	for node, attrs := range nodesAttrsMap {
		if userNode != nil && passwordNode != nil {
			//found both nodes
			return userNode, passwordNode, nil
		}
		val := attrs[config.PlaceHolderAttrName]
		if val == config.Config.UsernameInputPH {
			userNode = node
			continue
		}
		if val == config.Config.PasswordInputPH {
			passwordNode = node
			continue
		}
	}
	return userNode, passwordNode, fmt.Errorf("no node found has attributes: \"%s\" or \"%s\"", config.Config.UsernameInputPH, config.Config.PasswordInputPH)
}

func getLoginNode(ctx context.Context) (loginNode *cdp.Node, err error) {
	nodes, err := common.GetAllButtonNodes(ctx)
	if err != nil {
		return loginNode, fmt.Errorf("unable to get all button nodes: %+v", err)
	}

	nodesAttrsMap := common.GetNodesAttrsMap(nodes)
	for node, attrs := range nodesAttrsMap {
		typeVal := attrs[config.TypeAttrName]
		if typeVal != loginButtonTypeAttrVal {
			continue
		}
		var text string
		chromedp.Run(ctx,
			chromedp.Text(config.ButtonNodeSel, &text, chromedp.ByQuery, chromedp.FromNode(node.Parent)),
		)
		if text == config.Config.LoginButtonText {
			return node, nil
		}
	}
	return loginNode, fmt.Errorf("no button node found has type attr of \"%s\" and text of \"%s\"", loginButtonTypeAttrVal, config.Config.LoginButtonText)
}

func navigateToPixivSiteAndClickLogin(ctx context.Context) (err error) {
	return chromedp.Run(ctx,
		chromedp.Navigate(config.PixivSiteUrl),
		// wait for element is visible (ie, page is loaded)
		chromedp.WaitVisible(`a.signup-form__submit--login`),
		// // find and click
		chromedp.Click(`a.signup-form__submit--login`, chromedp.NodeVisible),
		// just wait
		chromedp.Sleep(3*time.Second),
	)
}

func loginPixiv(ctx context.Context, screenshotBuf *[]byte) (err error) {
	err = navigateToPixivSiteAndClickLogin(ctx)
	if err != nil {
		return fmt.Errorf("failed to navigate to pixiv login page: %+v", err)
	}

	userNode, pwNode, err := getUserAndPasswordInputNodes(ctx)
	if err != nil {
		return fmt.Errorf("unable to find input nodes of user or password: %+v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.SendKeys(config.InputNodeSel, config.Config.Username, chromedp.ByQuery, chromedp.FromNode(userNode.Parent)),
		chromedp.SendKeys(config.InputNodeSel, config.Config.Password, chromedp.ByQuery, chromedp.FromNode(pwNode.Parent)),
		// just wait
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		return fmt.Errorf("unable to send keys to username or password input: %+v", err)
	}

	loginNode, err := getLoginNode(ctx)
	if err != nil {
		return fmt.Errorf("unable to find node of login button: %+v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(config.ButtonNodeSel, chromedp.ByQuery, chromedp.FromNode(loginNode.Parent)),
		chromedp.WaitVisible(`img.sc-1yo2nn9-1.bBQkQw`),
		// just wait
		chromedp.Sleep(3*time.Second),
		// take screenshot
		chromedp.FullScreenshot(screenshotBuf, 90),
	)
	if err != nil {
		return fmt.Errorf("failed to click login button: %+v", err)
	}
	return nil
}
