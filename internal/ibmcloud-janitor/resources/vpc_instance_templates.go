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

type VPCInstanceTemplate struct{}

// Cleans up the instance templates in the target VPC.
func (VPCInstanceTemplate) cleanup(options *CleanupOptions) error {
	resourceLogger := logrus.WithFields(logrus.Fields{"resource": options.Resource.Name})
	resourceLogger.Info("Cleaning up the instance templates")
	client, err := NewVPCClient(options)
	if err != nil {
		return errors.Wrap(err, "couldn't create VPC client")
	}

	templateList, _, err := client.ListInstanceTemplates(&vpcv1.ListInstanceTemplatesOptions{})
	if err != nil {
		return errors.Wrap(err, "failed to list the instance templates")
	}

	for _, templateIntf := range templateList.Templates {
		template, ok := templateIntf.(*vpcv1.InstanceTemplate)
		if !ok {
			continue
		}
		if !instanceTemplateBelongsToVPC(template, client.VPCID) {
			continue
		}
		_, err := client.DeleteInstanceTemplate(&vpcv1.DeleteInstanceTemplateOptions{
			ID: template.ID,
		})
		if err != nil {
			return errors.Wrapf(err, "failed to delete the instance template %q", *template.Name)
		}
		resourceLogger.WithFields(logrus.Fields{"name": template.Name}).Info("Successfully deleted the instance template")
	}
	resourceLogger.Info("Successfully deleted the instance templates")
	return nil
}

func instanceTemplateBelongsToVPC(template *vpcv1.InstanceTemplate, vpcID string) bool {
	switch vpc := template.VPC.(type) {
	case *vpcv1.VPCIdentity:
		return vpc.ID != nil && *vpc.ID == vpcID
	case *vpcv1.VPCIdentityByID:
		return vpc.ID != nil && *vpc.ID == vpcID
	default:
		return false
	}
}
