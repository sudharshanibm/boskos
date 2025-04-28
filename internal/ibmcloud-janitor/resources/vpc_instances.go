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

	instanceListOpts := &vpcv1.ListInstancesOptions{}
	if client.VPCID != "" {
		instanceListOpts.VPCID = &client.VPCID
	} else {
		instanceListOpts.ResourceGroupID = &client.ResourceGroupID
	}

	instanceList, _, err := client.ListInstances(instanceListOpts)
	if err != nil {
		return errors.Wrap(err, "failed to list the instances")
	}

	for _, instance := range instanceList.Instances {
		_, err = client.DeleteInstance(&vpcv1.DeleteInstanceOptions{
			ID: instance.ID,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to delete the instance %q", *instance.Name)
		}
		resourceLogger.WithFields(logrus.Fields{"name": instance.Name}).Info("Successfully deleted instance")
	}

	resourceLogger.Info("Successfully deleted all virtual server instances")
	return nil
}
