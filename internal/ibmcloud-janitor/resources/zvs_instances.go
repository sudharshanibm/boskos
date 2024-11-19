package resources

import (
	"context"
	"github.com/IBM-Cloud/zvs-go-client/clients/vsi"
	"github.com/IBM-Cloud/zvs-go-client/zvssession"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ZVSIInstance struct {
	client *vsi.ZVSIClient
	serviceInstanceID string
}

func NewVSIClient(sess *zvssession.ZVSSession, instanceID string) *ZVSIInstance {
	client := &ZVSIInstance{
		serviceInstanceID: instanceID,
	}
	client.client = vsi.NewZVSIClient(context.Background(), sess, instanceID)
	return client
}

func (i ZVSIInstance) cleanup(options *CleanupOptions) error {
	resourceLogger := logrus.WithFields(logrus.Fields{"resource": options.Resource.Name})
	resourceLogger.Info("Cleaning up the VSI instances")
	zvsiClient, err := NewZVSIClient(options)
	if err != nil {
		return errors.Wrap(err, "couldn't create Z VSI client")
	}

	instances, err := zvsiClient.GetInstances()
	if err != nil {
		return errors.Wrapf(err, "failed to get the instances in %q", zvsiClient.resource.Name)
	}

	for _, ins := range instances.VSIInstances {
		err = zvsiClient.DeleteInstance(*ins.InstanceID)
		if err != nil {
			return errors.Wrapf(err, "failed to delete the instance %q", *ins.Name)
		}
	}
	resourceLogger.Info("Successfully deleted VSI instances")
	return nil
}
