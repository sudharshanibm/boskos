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
	"github.com/IBM/vpc-go-sdk/vpcv1"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type VPCInstance struct{}

// Cleans up the virtual server instances in a given region
func (VPCInstance) cleanup(options *CleanupOptions) error {
	resourceLogger := logrus.WithFields(logrus.Fields{"resource": options.Resource.Name})
	resourceLogger.Info("Cleaning up the virtual server instances")
	client, err := NewVPCClient(options)
	if err != nil {
		return errors.Wrap(err, "couldn't create VPC client")
	}

	listInstanceOpts := &vpcv1.ListInstancesOptions{
		VPCID: &client.VPCID,
	}

	instanceList, _, err := client.ListInstances(listInstanceOpts)
	if err != nil {
		return errors.Wrap(err, "failed to list the instances")
	}

	for _, ins := range instanceList.Instances {
		if err := deleteInstanceFloatingIPs(client, ins); err != nil {
			return err
		}
		_, err := client.DeleteInstance(&vpcv1.DeleteInstanceOptions{
			ID: ins.ID,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to delete the instance %q", *ins.Name)
		}
	}
	resourceLogger.Info("Successfully deleted the virtual server instances")
	return nil
}

func deleteInstanceFloatingIPs(client *IBMVPCClient, ins vpcv1.Instance) error {
	interfaces, _, err := client.ListInstanceNetworkInterfaces(&vpcv1.ListInstanceNetworkInterfacesOptions{
		InstanceID: ins.ID,
	})
	if err != nil {
		return errors.Wrapf(err, "failed to list network interfaces for instance %q", *ins.Name)
	}

	for _, networkInterface := range interfaces.NetworkInterfaces {
		for _, fip := range networkInterface.FloatingIps {
			_, err := client.DeleteFloatingIP(&vpcv1.DeleteFloatingIPOptions{
				ID: fip.ID,
			})
			if err != nil {
				return errors.Wrapf(err, "failed to delete the floating IP %q", *fip.Name)
			}
		}
	}

	return nil
}
