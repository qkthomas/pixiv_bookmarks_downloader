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
package download

import (
	"fmt"
	"sync"

	"github.com/chromedp/cdproto/network"
)

var (
	Manager = new(eventManager)
)

type eventManager struct {
	events      map[network.RequestID]Event
	eventsMutex sync.Mutex
}

func (manager *eventManager) RegisterEvent(requestID network.RequestID, eh EventHandle) {
	manager.eventsMutex.Lock()
	defer manager.eventsMutex.Unlock()
	if manager.events == nil {
		manager.events = make(map[network.RequestID]Event)
	}
	manager.events[requestID] = Event{
		id:     requestID,
		handle: eh,
	}
}

func (manager *eventManager) triggerEvent(requestID network.RequestID,
	errFuncWhenEventNotExist func(network.RequestID, Event) error) (err error) {
	manager.eventsMutex.Lock()
	defer manager.eventsMutex.Unlock()

	if errFuncWhenEventNotExist == nil {
		//avoid panic
		errFuncWhenEventNotExist = func(requestID network.RequestID, ev Event) error {
			return nil
		}
	}

	event, exist := manager.events[requestID]
	if !exist {
		return errFuncWhenEventNotExist(requestID, event)
	}
	var selfRemove bool
	defer func() {
		if selfRemove {
			delete(manager.events, requestID)
		}
	}()
	selfRemove, err = event.handle()
	return err
}

func (manager *eventManager) TriggerEvent(requestID network.RequestID) (err error) {
	errFuncWhenEventNotExist := func(requestID network.RequestID, ev Event) error {
		return fmt.Errorf("event with id \"%s\" does not exist", requestID.String())
	}
	return manager.triggerEvent(requestID, errFuncWhenEventNotExist)
}

func (manager *eventManager) TriggerEventIfExist(requestID network.RequestID) (err error) {
	errFuncWhenEventNotExist := func(requestID network.RequestID, ev Event) error {
		return nil
	}
	return manager.triggerEvent(requestID, errFuncWhenEventNotExist)
}
