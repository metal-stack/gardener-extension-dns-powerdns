// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
// SPDX-FileCopyrightText: 2024 metal-stack Authors
//
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"strings"

	"github.com/joeig/go-powerdns/v3"
)

// NewClient returns a new powerdns client from the given credentials.
func NewClient(cred *Credentials) *Client {
	headers := map[string]string{"X-API-Key": cred.ApiKey}
	httpClient := newHttpClient(cred.InsecureSkipVerify, cred.TrustedCaCert)
	virtualHost := "localhost"
	if cred.VirtualHost != nil {
		virtualHost = *cred.VirtualHost
	}

	return &Client{
		powerdns: powerdns.NewClient(cred.Server, virtualHost, headers, httpClient),
	}
}

func newHttpClient(insecureSkipVerify bool, trustedCaCert []byte) *http.Client {
	httpClient := http.DefaultClient

	if insecureSkipVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	if trustedCaCert != nil {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(trustedCaCert)
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:            caCertPool,
				InsecureSkipVerify: false,
				MinVersion:         tls.VersionTLS12,
			},
		}
	}

	return httpClient
}

// GetDNSHostedZones returns a map of all DNS hosted zone names mapped to their IDs.
func (c *Client) GetDNSHostedZones(ctx context.Context) (map[string]string, error) {
	zones := make(map[string]string)
	zones2, err := c.powerdns.Zones.List(ctx)
	if err != nil {
		return nil, err
	}
	for _, zone := range zones2 {
		zones[normalizeName(stringValue(zone.Name))] = stringValue(zone.ID)
	}
	return zones, nil
}

// CreateOrUpdateDNSRecordSet creates or updates the DNS recordset in the DNS hosted zone with the given zone ID, name, type, values, and TTL.
func (c *Client) CreateOrUpdateDNSRecordSet(ctx context.Context, zoneId, name string, recordType powerdns.RRType, values []string, ttl int64) error {
	return c.powerdns.Records.Change(ctx, zoneId, name, recordType, uint32(ttl), values)
}

// DeleteDNSRecordSet deletes the DNS recordset in the DNS hosted zone with the given zone ID, name, type.
func (c *Client) DeleteDNSRecordSet(ctx context.Context, zoneId, name string, recordType powerdns.RRType) error {
	return c.powerdns.Records.Delete(ctx, zoneId, name, recordType)
}

func stringValue(v *string) string {
	if v != nil {
		return *v
	}
	return ""
}

func normalizeName(name string) string {
	if strings.HasPrefix(name, "\\052.") {
		name = "*" + name[4:]
	}
	if strings.HasSuffix(name, ".") {
		return name[:len(name)-1]
	}
	return name
}
