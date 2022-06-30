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
	submitButtonTypeAttrVal = `submit`
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

//get the login button on the page that you input username and password
func getSubmitLoginNode(ctx context.Context) (submitLoginNode *cdp.Node, err error) {
	return getSubmitButtonNode(ctx, config.Config.LoginButtonText)
}

func navigateToPixivSiteAndClickLogin(ctx context.Context) (err error) {
	return chromedp.Run(ctx,
		chromedp.Navigate(config.PixivSiteUrl),
		// wait for element is visible (ie, page is loaded)
		chromedp.WaitVisible(`a.signup-form__submit--login`, chromedp.ByQuery),
		// // find and click
		chromedp.Click(`a.signup-form__submit--login`, chromedp.ByQuery, chromedp.NodeVisible),
		// just wait
		chromedp.Sleep(3*time.Second),
	)
}

func loginPixiv(ctx context.Context) (err error) {
	err = navigateToPixivSiteAndClickLogin(ctx)
	if err != nil {
		return fmt.Errorf("failed to navigate to pixiv login page: %+v", err)
	}

	userNode, pwNode, err := getUserAndPasswordInputNodes(ctx)
	if err != nil {
		return fmt.Errorf("unable to find input nodes of user or password: %+v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.SendKeys(config.AnySel, config.Config.Username, chromedp.ByQuery, common.TargetNode(userNode)),
		chromedp.SendKeys(config.AnySel, config.Config.Password, chromedp.ByQuery, common.TargetNode(pwNode)),
		// just wait
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		return fmt.Errorf("unable to send keys to username or password input: %+v", err)
	}

	loginNode, err := getSubmitLoginNode(ctx)
	if err != nil {
		return fmt.Errorf("unable to find node of login button: %+v", err)
	}

	err = chromedp.Run(ctx,
		chromedp.Click(config.AnySel, chromedp.ByQuery, common.TargetNode(loginNode)),
		chromedp.WaitVisible(config.TopLeftPixivImgSel, chromedp.ByQuery),
		// just wait
		chromedp.Sleep(3*time.Second),
	)
	if err != nil {
		return fmt.Errorf("failed to click login button: %+v", err)
	}
	return nil
}
