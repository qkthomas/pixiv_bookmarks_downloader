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
	"fmt"
	"strings"
	"sync"
)

type Errors struct {
	lock   sync.Mutex
	errors []error
}

// Add adds a new error
func (c *Errors) Add(err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if err == nil {
		return
	}
	c.errors = append(c.errors, err)
}

// Get gets all error in the struct
func (c *Errors) Get() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	return ConcatenateErrors(c.errors...)
}

// Get gets all error in the struct
func (c *Errors) Size() int {
	c.lock.Lock()
	defer c.lock.Unlock()

	return len(c.errors)
}

// Clean cleans list of errors
func (c *Errors) Clean() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.errors = nil
}

func ConcatenateErrors(errors ...error) error {
	var errorSlice []string
	for _, err := range errors {
		if err != nil {
			errorSlice = append(errorSlice, err.Error())
		}
	}
	if len(errorSlice) == 0 {
		return nil
	}

	return fmt.Errorf(strings.Join(errorSlice, "\n -"))
}
