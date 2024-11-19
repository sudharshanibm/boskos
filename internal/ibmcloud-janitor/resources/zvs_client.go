package resources

import (
	"github.com/IBM-Cloud/zvs-go-client/zvssession"
	"github.com/IBM-Cloud/zvs-go-client/zvsmodels"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type ZVSIClient struct {
	session *zvssession.ZVSSession
	instance *ZVSIInstance
	network  *ZVSINetwork
	resource *common.Resource
}

func (z *ZVSIClient) GetInstances() (*zvsmodels.VSIInstances, error) {
	return z.instance.GetAll()
}

func (z *ZVSIClient) DeleteInstance(id string) error {
	return z.instance.Delete(id)
}

func NewZVSIClient(options *CleanupOptions) (*ZVSIClient, error) {
	resourceLogger := logrus.WithFields(logrus.Fields{"resource": options.Resource.Name})
	client := &ZVSIClient{}
	zvsData, err := ibmcloud.GetZVSIResourceData(options.Resource)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the resource data")
	}

	auth, err := account.GetAuthenticator()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get the authenticator")
	}

	clientOptions := &zvssession.ZVSOptions{
		Debug:         options.Debug,
		Authenticator: auth,
		Zone:          zvsData.Zone,
	}
	client.session, err = zvssession.NewZVSSession(clientOptions)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create a new session")
	}
	resourceLogger.Info("successfully created Z VSI client")

	client.instance = NewVSIClient(client.session, zvsData.ServiceInstanceID)
	client.network = NewVSINetworkClient(client.session, zvsData.ServiceInstanceID)
	client.resource = options.Resource

	return client, nil
}
