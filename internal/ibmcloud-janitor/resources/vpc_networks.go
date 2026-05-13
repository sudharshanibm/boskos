/*
Copyright 2022 The Kubernetes Authors.

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

package resources

import (
	"strings"
	"time"

	"github.com/IBM/vpc-go-sdk/vpcv1"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/wait"
)

type VPCNetwork struct{}

var (
	fipDeletionTimeout  = time.Minute * 4
	fipPollingInterval  = time.Second * 15
	fipNotFoundPatterns = []string{"cannot be found", "not found"}
)

// Clean up of network resources is done in the following order:
// 1. Unset and delete public gateways attached to a subnet
// 2. Delete the subnet
// 3. Delete any remaining public gateways in the target VPC
func (VPCNetwork) cleanup(options *CleanupOptions) error {
	resourceLogger := logrus.WithFields(logrus.Fields{"resource": options.Resource.Name})
	resourceLogger.Info("Cleaning up the networks")
	client, err := NewVPCClient(options)
	if err != nil {
		return errors.Wrap(err, "couldn't create VPC client")
	}

	listSubnetOpts := &vpcv1.ListSubnetsOptions{
		VPCID: &client.VPCID,
	}

	subnetList, _, err := client.ListSubnets(listSubnetOpts)
	if err != nil {
		return errors.Wrap(err, "failed to list the subnets")
	}

	for _, subnet := range subnetList.Subnets {
		pg, _, err := client.GetSubnetPublicGateway(&vpcv1.GetSubnetPublicGatewayOptions{
			ID: subnet.ID,
		})
		if pg != nil && err == nil {
			floatingIPID := ""
			if pg.FloatingIP != nil && pg.FloatingIP.ID != nil {
				floatingIPID = *pg.FloatingIP.ID
			}
			_, err := client.UnsetSubnetPublicGateway(&vpcv1.UnsetSubnetPublicGatewayOptions{
				ID: subnet.ID,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to unset the gateway for %q", *subnet.Name)
			}

			_, err = client.DeletePublicGateway(&vpcv1.DeletePublicGatewayOptions{
				ID: pg.ID,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to delete the gateway %q", *pg.Name)
			}
			resourceLogger.WithFields(logrus.Fields{"name": pg.Name}).Info("Successfully deleted the gateway")
			if floatingIPID != "" {
				if err := deleteFloatingIP(client, floatingIPID, resourceLogger); err != nil {
					return err
				}
			}
		}
		_, err = client.DeleteSubnet(&vpcv1.DeleteSubnetOptions{ID: subnet.ID})
		if err != nil {
			return errors.Wrapf(err, "failed to delete the subnet %q", *subnet.Name)
		}
	}

	if err := deletePublicGateways(client, resourceLogger); err != nil {
		return err
	}

	resourceLogger.Info("Successfully deleted VPC network resources")
	return nil
}

func deletePublicGateways(client *IBMVPCClient, resourceLogger *logrus.Entry) error {
	publicGateways, _, err := client.ListPublicGateways(&vpcv1.ListPublicGatewaysOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list the public gateways")
	}

	for _, pg := range publicGateways.PublicGateways {
		if pg.VPC == nil || pg.VPC.ID == nil || *pg.VPC.ID != client.VPCID {
			continue
		}
		floatingIPID := ""
		if pg.FloatingIP != nil && pg.FloatingIP.ID != nil {
			floatingIPID = *pg.FloatingIP.ID
		}
		_, err := client.DeletePublicGateway(&vpcv1.DeletePublicGatewayOptions{
			ID: pg.ID,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to delete the gateway %q", *pg.Name)
		}
		resourceLogger.WithFields(logrus.Fields{"name": pg.Name}).Info("Successfully deleted the gateway")
		if floatingIPID != "" {
			if err := deleteFloatingIP(client, floatingIPID, resourceLogger); err != nil {
				return err
			}
		}
	}

	return nil
}

func deleteFloatingIP(client *IBMVPCClient, id string, resourceLogger *logrus.Entry) error {
	var lastErr error
	err := wait.PollImmediate(fipPollingInterval, fipDeletionTimeout, func() (bool, error) {
		_, err := client.DeleteFloatingIP(&vpcv1.DeleteFloatingIPOptions{
			ID: &id,
		})
		if err == nil {
			return true, nil
		}
		if isFloatingIPNotFound(err) {
			return true, nil
		}
		lastErr = err
		return false, nil
	})
	if err != nil {
		return errors.Wrapf(lastErr, "failed to delete the floating IP %q", id)
	}
	resourceLogger.WithFields(logrus.Fields{"id": id}).Info("Successfully deleted the floating IP")
	return nil
}

func isFloatingIPNotFound(err error) bool {
	for _, pattern := range fipNotFoundPatterns {
		if strings.Contains(err.Error(), pattern) {
			return true
		}
	}
	return false
}
