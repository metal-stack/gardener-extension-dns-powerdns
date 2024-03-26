// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
// SPDX-FileCopyrightText: 2024 metal-stack Authors
//
// SPDX-License-Identifier: Apache-2.0

package powerdns

import (
	"context"
	"fmt"
	"strconv"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	pdnsclient "github.com/metal-stack/gardener-extension-dns-powerdns/pkg/powerdns/client"
)

// NewClientFromSecretRef creates a new Client from given k8s <secretRef>
func NewClientFromSecretRef(ctx context.Context, client client.Client, secretRef corev1.SecretReference) (*pdnsclient.Client, error) {
	credentials, err := getCredentialsFromSecretRef(ctx, client, secretRef)
	if err != nil {
		return nil, err
	}
	return pdnsclient.NewClient(credentials), nil
}

func getCredentialsFromSecretRef(ctx context.Context, client client.Client, secretRef corev1.SecretReference) (*pdnsclient.Credentials, error) {
	secret, err := extensionscontroller.GetSecretByReference(ctx, client, &secretRef)
	if err != nil {
		return nil, err
	}
	return readCredentialsSecret(secret)
}

func readCredentialsSecret(secret *corev1.Secret) (*pdnsclient.Credentials, error) {
	if secret.Data == nil {
		return nil, fmt.Errorf("secret does not contain any data")
	}

	apiKey, err := getSecretStringValue(secret, ApiKey, true)
	if err != nil {
		return nil, err
	}

	server, _ := getSecretStringValue(secret, Server, true)
	if err != nil {
		return nil, err
	}

	virtualHost, err := getSecretStringValue(secret, VirtualHost, false)

	insecureSkipVerify, err := getSecretBoolValue(secret, InsecureSkipVerify)
	if err != nil {
		return nil, err
	}

	trustedCaCert, _ := getSecretValue(secret, TrustedCaCert, false)

	return &pdnsclient.Credentials{
		ApiKey:             *apiKey,
		Server:             *server,
		VirtualHost:        virtualHost,
		InsecureSkipVerify: insecureSkipVerify,
		TrustedCaCert:      trustedCaCert,
	}, nil
}

func getSecretValue(secret *corev1.Secret, key string, required bool) ([]byte, error) {
	if value, ok := secret.Data[key]; ok {
		return value, nil
	}
	if required {
		return nil, fmt.Errorf("missing %q field in secret", key)
	}
	return nil, nil
}

func getSecretStringValue(secret *corev1.Secret, key string, required bool) (*string, error) {
	value, err := getSecretValue(secret, key, required)
	if err != nil {
		return nil, err
	}
	return ptr.To(string(value)), nil
}

func getSecretBoolValue(secret *corev1.Secret, key string) (bool, error) {
	if raw, ok := secret.Data[key]; ok {
		value, err := strconv.ParseBool(string(raw))
		if err != nil {
			return false, fmt.Errorf("cannot parse %q field in secret: %w", key, err)
		}
		return value, nil
	}
	return false, nil
}
