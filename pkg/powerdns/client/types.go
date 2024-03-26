// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
// SPDX-FileCopyrightText: 2024 metal-stack Authors
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"github.com/joeig/go-powerdns/v3"
)

type Client struct {
	powerdns *powerdns.Client
}

type Credentials struct {
	ApiKey             string
	Server             string
	VirtualHost        *string
	InsecureSkipVerify bool
	TrustedCaCert      []byte
}
