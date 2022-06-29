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
package config

import (
	"regexp"
	"time"
)

const (
	//some urls
	PixivSiteUrl = `https://www.pixiv.net`

	//some selectors
	ThumbnailNodeSel   = `img.sc-rp5asc-10.erYaF`
	FigureNodeSel      = `figure.sc-1yvhotl-3.jUCdwp`
	TopLeftPixivImgSel = `img.sc-1yo2nn9-1.bBQkQw`
	InputNodeSel       = `input`
	ButtonNodeSel      = `button`
	ImgNodeSel         = `img`
	AnchorNodeSel      = `a`
	SvgNodeSel         = `svg`
	NavvNodeSel        = `nav`
	DivNodeSel         = `div`

	//some attribute names/keys
	PlaceHolderAttrName  = `placeholder`
	TypeAttrName         = `type`
	RelAttrName          = `rel`
	HrefAttrName         = `href`
	ClassAttrName        = `class`
	SrcAttrName          = `src`
	AltAttrName          = `alt`
	AriaDisabledAttrName = `aria-disabled`

	//some directory names
	SavedFileLocation      = `saved`
	ThumbnailsFileLocation = `thumbnails`

	ErrorMsgPrefix = `error:`
	InfMsgPrefix   = `info:`

	//time and duration
	SavingRespWaitDura = time.Second * 6

	//some file permission
	WriteFilePermission = 0644

	//some regex
	artworkerImgReStr              = `(\d+)_p(\d+)\.(jpg|png|jpeg|gif)+` //this only match full res img
	artworkIDReStr                 = `(\d+)_p`                           //this also match thumbnails
	userProfileImgSrcReStr         = `https:\/\/i\.pximg\.net\/user-profile\/img\/(.+\.jpg)`
	userPookmarkPageUrlSuffixReStr = `/users\/(\d+)\/bookmarks\/artworks\?*(p=(\d+))*`
)

var (
	ArtworkImgRe                = regexp.MustCompile(artworkerImgReStr)
	ArtworkIDRe                 = regexp.MustCompile(artworkIDReStr)
	UserProfileImgSrcRe         = regexp.MustCompile(userProfileImgSrcReStr)
	UserBookmarkPageUrSuffixlRe = regexp.MustCompile(userPookmarkPageUrlSuffixReStr)
)
