package client

import (
	"net/http"

	"git.prosoftke.sk/nghis/openapi/clients/go/nghisclinicalclient/v2"
	"github.com/arfis/waiting-room/nghis-adapter/internal/config/service"
)

// NewClinicalClient prepares client for the 'clinical' service
func NewClinicalClient(configuration *service.Configuration, httpClient *http.Client) *nghisclinicalclient.APIClient {
	clientCfg := nghisclinicalclient.Configuration{
		Host:          configuration.ClinicalClientHost,
		Scheme:        configuration.ClinicalClientScheme,
		DefaultHeader: map[string]string{},
		Debug:         false,
		UserAgent:     "",
		Servers: nghisclinicalclient.ServerConfigurations{
			nghisclinicalclient.ServerConfiguration{
				URL: configuration.ClinicalClientContext,
			},
		},
		OperationServers: map[string]nghisclinicalclient.ServerConfigurations{},
		HTTPClient:       httpClient,
	}

	return nghisclinicalclient.NewAPIClient(&clientCfg)
}
