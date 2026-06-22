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
	"github.com/IBM/go-sdk-core/v5/core"
	"github.com/IBM/platform-services-go-sdk/resourcemanagerv2"
	"github.com/pkg/errors"
)

// resourceGroupIDLength is the length of an IBM Cloud resource group ID, a
// 32-character hexadecimal string with no separators.
const resourceGroupIDLength = 32

// looksLikeResourceGroupID reports whether value is already an IBM Cloud
// resource group ID (32 hexadecimal characters) rather than a human-readable
// resource group name. It lets the resolver accept either form, so a name key
// and a legacy ID key can coexist during rollout without a flag day.
func looksLikeResourceGroupID(value string) bool {
	if len(value) != resourceGroupIDLength {
		return false
	}
	for _, c := range value {
		switch {
		case c >= '0' && c <= '9', c >= 'a' && c <= 'f', c >= 'A' && c <= 'F':
		default:
			return false
		}
	}
	return true
}

// resolveResourceGroupID returns the resource group ID for the given value.
// If value already looks like a resource group ID it is returned unchanged;
// otherwise it is treated as a resource group name and resolved to its ID via
// the Resource Manager API using auth and accountID. accountID may be nil, in
// which case the account scoped to the API key is used.
func resolveResourceGroupID(auth core.Authenticator, value string, accountID *string) (string, error) {
	if value == "" {
		return "", errors.New("empty resource group name")
	}
	if looksLikeResourceGroupID(value) {
		return value, nil
	}

	resourceManager, err := resourcemanagerv2.NewResourceManagerV2(&resourcemanagerv2.ResourceManagerV2Options{
		Authenticator: auth,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to create resource manager client")
	}

	listOptions := &resourcemanagerv2.ListResourceGroupsOptions{}
	listOptions.SetName(value)
	if accountID != nil {
		listOptions.SetAccountID(*accountID)
	}

	groups, _, err := resourceManager.ListResourceGroups(listOptions)
	if err != nil {
		return "", errors.Wrapf(err, "failed to list resource groups for name %q", value)
	}
	if groups == nil || len(groups.Resources) == 0 || groups.Resources[0].ID == nil {
		return "", errors.Errorf("no resource group found with name %q", value)
	}

	return *groups.Resources[0].ID, nil
}
