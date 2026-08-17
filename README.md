# Gardener Extension for PowerDNS

[![GitHub License](https://img.shields.io/github/license/metal-stack/gardener-extension-dns-powerdns)](https://github.com/metal-stack/gardener-extension-dns-powerdns/blob/main/LICENSE)
[![Build](https://github.com/metal-stack/gardener-extension-dns-powerdns/actions/workflows/build.yaml/badge.svg)](https://github.com/metal-stack/gardener-extension-dns-powerdns/actions/workflows/build.yaml)

[Project Gardener](https://gardener.cloud/) implements the automated management and operation of [Kubernetes](https://kubernetes.io/) clusters as a service. This controller implements Gardener's extension contract for the **PowerDNS** provider.

It reconciles the `DNSRecords` resources of `type: powerdns` against a [PowerDNS](https://www.powerdns.com/) server.

For more detailed documentation about the extension contract, please refer to the [Gardener docs](https://github.com/gardener/gardener/blob/master/docs/extensions/overview.md).

## Example

An example `ControllerRegistration` resource that can be used to register this controller to Gardener can be found [here](example/controller-registration.yaml).

## Development

The extension can be developed locally in the [mini-lab](https://github.com/metal-stack/mini-lab), which contains a working deployment of PowerDNS.

## Feedback and Support

Feedback and contributions are always welcome! Please report bugs or suggestions as [GitHub issues](https://github.com/metal-stack/gardener-extension-dns-powerdns/issues) or reach out to our [community](https://metal-stack.io/community).
