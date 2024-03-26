// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
// SPDX-FileCopyrightText: 2024 metal-stack Authors
//
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	controllercmd "github.com/gardener/gardener/extensions/pkg/controller/cmd"
	extensionsdnsrecordcontroller "github.com/gardener/gardener/extensions/pkg/controller/dnsrecord"
	extensionsheartbeatcontroller "github.com/gardener/gardener/extensions/pkg/controller/heartbeat"

	dnsrecordcontroller "github.com/metal-stack/gardener-extension-dns-powerdns/pkg/controller/dnsrecord"
)

// ControllerSwitchOptions are the controllercmd.SwitchOptions for the provider controllers.
func ControllerSwitchOptions() *controllercmd.SwitchOptions {
	return controllercmd.NewSwitchOptions(
		controllercmd.Switch(extensionsdnsrecordcontroller.ControllerName, dnsrecordcontroller.AddToManager),
		controllercmd.Switch(extensionsheartbeatcontroller.ControllerName, extensionsheartbeatcontroller.AddToManager),
	)
}
