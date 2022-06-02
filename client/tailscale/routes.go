// Copyright (c) 2022 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build go1.18
// +build go1.18

package tailscale

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"inet.af/netaddr"
)

// Routes contains the lists of subnet routes that are currently advertised by a device,
// as well as the subnets that are enabled to be routed by the device.
type Routes struct {
	AdvertisedRoutes []netaddr.IPPrefix `json:"advertisedRoutes"`
	EnabledRoutes    []netaddr.IPPrefix `json:"enabledRoutes"`
}

// Routes retrieves the list of subnet routes that have been enabled for a device.
// The routes that are returned are not necessarily advertised by the device,
// they have only been preapproved.
func (c *Client) Routes(ctx context.Context, deviceID string) (routes *Routes, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("tailscale.Routes: %w", err)
		}
	}()

	path := fmt.Sprintf("%s/api/v2/device/%s/routes", c.baseURL(), deviceID)
	req, err := http.NewRequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	b, resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	// If status code was not successful, return the error.
	// TODO: Change the check for the StatusCode to include other 2XX success codes.
	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(b, resp)
	}

	var sr Routes
	err = json.Unmarshal(b, &sr)
	return &sr, err
}

type postRoutesParams struct {
	Routes []netaddr.IPPrefix `json:"routes"`
}

// SetRoutes updates the list of subnets that are enabled for a device.
// Subnets must be parsable by inet.af/netaddr.ParseIPPrefix.
// Subnets do not have to be currently advertised by a device, they may be pre-enabled.
// Returns the updated list of enabled and advertised subnet routes in a *Routes object.
func (c *Client) SetRoutes(ctx context.Context, deviceID string, subnets []netaddr.IPPrefix) (routes *Routes, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("tailscale.SetRoutes: %w", err)
		}
	}()
	params := &postRoutesParams{Routes: subnets}
	data, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("%s/api/v2/device/%s/routes", c.baseURL(), deviceID)
	req, err := http.NewRequestWithContext(ctx, "POST", path, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	b, resp, err := c.sendRequest(req)
	if err != nil {
		return nil, err
	}
	// If status code was not successful, return the error.
	// TODO: Change the check for the StatusCode to include other 2XX success codes.
	if resp.StatusCode != http.StatusOK {
		return nil, handleErrorResponse(b, resp)
	}

	var srr *Routes
	if err := json.Unmarshal(b, &srr); err != nil {
		return nil, err
	}
	return srr, err
}