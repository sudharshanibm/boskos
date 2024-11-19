package resources

import (
	"context"
	"github.com/IBM-Cloud/zvs-go-client/clients/network"
	"github.com/IBM-Cloud/zvs-go-client/zvssession"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ZVSINetwork struct {
	client *network.ZVSINetworkClient
	serviceInstanceID string
}

func NewVSINetworkClient(sess *zvssession.ZVSSession, instanceID string) *ZVSINetwork {
	client := &ZVSINetwork{
		serviceInstanceID: instanceID,
	}
	client.client = network.NewZVSINetworkClient(context.Background(), sess, instanceID)
	return client
}

func (n ZVSINetwork) cleanup(options *CleanupOptions) error {
	resourceLogger := logrus.WithFields(logrus.Fields{"resource": options.Resource.Name})
	resourceLogger.Info("Cleaning up the VSI networks")
	zvsiClient, err := NewZVSIClient(options)
	if err != nil {
		return errors.Wrap(err, "couldn't create Z VSI client")
	}

	networks, err := zvsiClient.GetNetworks()
	if err != nil {
		return errors.Wrapf(err, "failed to get the networks in %q", zvsiClient.resource.Name)
	}

	for _, net := range networks.Networks {
		ports, err := zvsiClient.GetPorts(*net.NetworkID)
		if err != nil {
			return errors.Wrapf(err, "failed to get ports of network %q", *net.Name)
		}
		for _, port := range ports.Ports {
			err = zvsiClient.DeletePort(*net.NetworkID, *port.PortID)
			if err != nil {
				return errors.Wrapf(err, "failed to delete port of network %q", *net.Name)
			}
		}
		err = zvsiClient.DeleteNetwork(*net.NetworkID)
		if err != nil {
			return errors.Wrapf(err, "failed to delete network %q", *net.Name)
		}
	}
	resourceLogger.Info("Successfully deleted the networks")
	return nil
}
