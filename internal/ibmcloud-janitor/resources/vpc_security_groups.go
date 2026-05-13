/*
Copyright 2026 The Kubernetes Authors.

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

type VPCSecurityGroup struct{}

// Cleans up non-default security groups in the target VPC.
func (VPCSecurityGroup) cleanup(options *CleanupOptions) error {
	resourceLogger := logrus.WithFields(logrus.Fields{"resource": options.Resource.Name})
	resourceLogger.Info("Cleaning up the security groups")
	client, err := NewVPCClient(options)
	if err != nil {
		return errors.Wrap(err, "couldn't create VPC client")
	}

	securityGroups, _, err := client.ListSecurityGroups(&vpcv1.ListSecurityGroupsOptions{
		VPCID: &client.VPCID,
	})
	if err != nil {
		return errors.Wrap(err, "failed to list the security groups")
	}

	for _, securityGroup := range securityGroups.SecurityGroups {
		if securityGroup.ID == nil {
			continue
		}
		if client.DefaultSecurityGroupID != "" && *securityGroup.ID == client.DefaultSecurityGroupID {
			continue
		}
		if len(securityGroup.Targets) != 0 {
			resourceLogger.WithFields(logrus.Fields{
				"name": securityGroup.Name,
				"id":   securityGroup.ID,
			}).Warn("Skipping security group because it still has targets")
			continue
		}

		_, err := client.DeleteSecurityGroup(&vpcv1.DeleteSecurityGroupOptions{
			ID: securityGroup.ID,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to delete the security group %q", *securityGroup.Name)
		}
		resourceLogger.WithFields(logrus.Fields{"name": securityGroup.Name}).Info("Successfully deleted the security group")
	}

	resourceLogger.Info("Successfully deleted the security groups")
	return nil
}
