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

	"github.com/google/uuid"
)

var (
	Manager = new(eventManager)
)

type eventManager struct {
	events map[uuid.UUID]Event
}

func (manager *eventManager) RegisterEvent(eh EventHandle) (id uuid.UUID) {
	id = uuid.New()
	for {
		_, duplicate := manager.events[id]
		if !duplicate {
			//the id is ok
			break
		}
		//to generate a new id
		id = uuid.New()
	}
	manager.events[id] = Event{
		id:     id,
		handle: eh,
	}
	return id
}

func (manager *eventManager) TriggerEvent(id uuid.UUID) (err error) {
	event, exist := manager.events[id]
	if !exist {
		return fmt.Errorf("event with id \"%s\" does not exist", id.String())
	}
	var selfRemove bool
	defer func() {
		if selfRemove {
			delete(manager.events, id)
		}
	}()
	selfRemove, err = event.handle()
	if err != nil {
		return err
	}
	return nil
}
