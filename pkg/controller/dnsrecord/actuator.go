// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
// SPDX-FileCopyrightText: 2024 metal-stack Authors
//
// SPDX-License-Identifier: Apache-2.0

package dnsrecord

import (
	"context"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/dnsrecord"
	gardencorev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	gardencorev1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	extensionsv1alpha1helper "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1/helper"
	kutil "github.com/gardener/gardener/pkg/utils/kubernetes"
	"github.com/go-logr/logr"
	"github.com/joeig/go-powerdns/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	pdns "github.com/metal-stack/gardener-extension-dns-powerdns/pkg/powerdns"
	pdnsclient "github.com/metal-stack/gardener-extension-dns-powerdns/pkg/powerdns/client"
)

type actuator struct {
	client client.Client
}

// NewActuator creates a new dnsrecord.Actuator.
func NewActuator(mgr manager.Manager) dnsrecord.Actuator {
	return &actuator{
		client: mgr.GetClient(),
	}
}

// Reconcile reconciles the DNSRecord.
func (a *actuator) Reconcile(ctx context.Context, log logr.Logger, dns *extensionsv1alpha1.DNSRecord, _ *extensionscontroller.Cluster) error {
	// Create PowerDNS client
	pdnsClient, err := pdns.NewClientFromSecretRef(ctx, a.client, dns.Spec.SecretRef)
	if err != nil {
		return fmt.Errorf("could not create PowerDNS client: %+v", err)
	}

	// Determine DNS hosted zone ID
	zone, err := a.getZone(ctx, log, dns, pdnsClient)
	if err != nil {
		return err
	}

	// Create or update DNS recordset
	ttl := extensionsv1alpha1helper.GetDNSRecordTTL(dns.Spec.TTL)
	log.Info("Creating or updating DNS recordset", "zone", zone, "name", dns.Spec.Name, "type", dns.Spec.RecordType, "values", dns.Spec.Values, "dnsrecord", kutil.ObjectName(dns))
	if err := pdnsClient.CreateOrUpdateDNSRecordSet(ctx, zone, dns.Spec.Name, powerdns.RRType(dns.Spec.RecordType), dns.Spec.Values, ttl); err != nil {
		return fmt.Errorf("could not create or update DNS recordset in zone %s with name %s, type %s, and values %v: %w", zone, dns.Spec.Name, dns.Spec.RecordType, dns.Spec.Values, err)
	}

	// Delete meta DNS recordset if exists
	if dns.Status.LastOperation == nil || dns.Status.LastOperation.Type == gardencorev1beta1.LastOperationTypeCreate {
		name, recordType := dnsrecord.GetMetaRecordName(dns.Spec.Name), "TXT"
		log.Info("Deleting meta DNS recordset", "zone", zone, "name", name, "type", recordType, "dnsrecord", kutil.ObjectName(dns))
		if err := pdnsClient.DeleteDNSRecordSet(ctx, zone, name, powerdns.RRType(recordType)); err != nil {
			return fmt.Errorf("could not delete meta DNS recordset in zone %s with name %s and type %s: %w", zone, name, recordType, err)
		}
	}

	// Update resource status
	patch := client.MergeFrom(dns.DeepCopy())
	dns.Status.Zone = &zone
	return a.client.Status().Patch(ctx, dns, patch)
}

// Delete deletes the DNSRecord.
func (a *actuator) Delete(ctx context.Context, log logr.Logger, dns *extensionsv1alpha1.DNSRecord, _ *extensionscontroller.Cluster) error {
	// Create PowerDNS client
	pdnsClient, err := pdns.NewClientFromSecretRef(ctx, a.client, dns.Spec.SecretRef)
	if err != nil {
		return fmt.Errorf("could not create PowerDNS client: %+v", err)
	}

	// Determine DNS hosted zone ID
	zone, err := a.getZone(ctx, log, dns, pdnsClient)
	if err != nil {
		return err
	}

	// Delete DNS recordset
	log.Info("Deleting DNS recordset", "zone", zone, "name", dns.Spec.Name, "type", dns.Spec.RecordType, "values", dns.Spec.Values, "dnsrecord", kutil.ObjectName(dns))
	if err := pdnsClient.DeleteDNSRecordSet(ctx, zone, dns.Spec.Name, powerdns.RRType(dns.Spec.RecordType)); err != nil {
		return fmt.Errorf("could not delete DNS recordset in zone %s with name %s, type %s, and values %v: %w", zone, dns.Spec.Name, dns.Spec.RecordType, dns.Spec.Values, err)
	}

	return nil
}

// Restore restores the DNSRecord.
func (a *actuator) Restore(ctx context.Context, log logr.Logger, dns *extensionsv1alpha1.DNSRecord, cluster *extensionscontroller.Cluster) error {
	return a.Reconcile(ctx, log, dns, cluster)
}

// Migrate migrates the DNSRecord.
func (a *actuator) Migrate(_ context.Context, _ logr.Logger, _ *extensionsv1alpha1.DNSRecord, _ *extensionscontroller.Cluster) error {
	return nil
}

func (a *actuator) getZone(ctx context.Context, log logr.Logger, dns *extensionsv1alpha1.DNSRecord, pdnsClient *pdnsclient.Client) (string, error) {
	switch {
	case dns.Spec.Zone != nil && *dns.Spec.Zone != "":
		return *dns.Spec.Zone, nil
	case dns.Status.Zone != nil && *dns.Status.Zone != "":
		return *dns.Status.Zone, nil
	default:
		// The zone is not specified in the resource status or spec. Try to determine the zone by
		// getting all hosted zones of the account and searching for the longest zone name that is a suffix of dns.spec.Name
		zones, err := pdnsClient.GetDNSHostedZones(ctx)
		if err != nil {
			return "", fmt.Errorf("could not get DNS hosted zones: %w", err)
		}
		log.Info("Got DNS hosted zones", "zones", zones, "dnsrecord", kutil.ObjectName(dns))
		zone := dnsrecord.FindZoneForName(zones, dns.Spec.Name)
		if zone == "" {
			return "", gardencorev1beta1helper.NewErrorWithCodes(fmt.Errorf("could not find DNS hosted zone for name %s", dns.Spec.Name), gardencorev1beta1.ErrorConfigurationProblem)
		}
		return zone, nil
	}
}
