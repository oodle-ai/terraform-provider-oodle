// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build !go1.6 && go1.5
// +build !go1.6,go1.5

package session

import (
	"net"
	"net/http"
	"time"
)

// Transport that should be used when a custom CA bundle is specified with the
// SDK.
func getCustomTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}
