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
	"regexp"
)

func Get1stGroupMatch(str string, re *regexp.Regexp) string {
	return getGroupMatch(str, re, 1)
}

func Get2ndGroupMatch(str string, re *regexp.Regexp) string {
	return getGroupMatch(str, re, 2)
}

func Get3rdGroupMatch(str string, re *regexp.Regexp) string {
	return getGroupMatch(str, re, 3)
}

func getGroupMatch(str string, re *regexp.Regexp, groupIdx int) string {
	matches := re.FindStringSubmatch(str)
	if len(matches) < re.NumSubexp()+1 {
		// no match
		return ""
	}
	return matches[groupIdx]
}
