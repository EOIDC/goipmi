/*
Copyright (c) 2014 VMware, Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package ipmi

import "fmt"

type transport interface {
	open() error
	close() error
	send(*Request, Response) error
	// Console enters Serial Over LAN mode
	Console() error
}

func newTransport(c *Connection) (transport, error) {
	switch c.Interface {
	case "lan":
		if c.Path == "" {
			return newLanTransport(c), nil
		}
		return newToolTransport(c), nil
	case "lanplus":
		return newToolTransport(c), nil
	case "open":
		return newOpenTransport(c), nil
	default:
		return nil, fmt.Errorf("unsupported interface: %s", c.Interface)
	}
}
