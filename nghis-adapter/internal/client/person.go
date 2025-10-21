package client

import (
	"fmt"
	"net/http"

	"git.prosoftke.sk/nghis/openapi/clients/go/nghispersonserviceclient"
	"github.com/arfis/waiting-room/nghis-adapter/internal/config/service"
	appCtx "github.com/arfis/waiting-room/nghis-adapter/internal/context"
)

// NewPersonClient prepares client for the 'person' service
func NewPersonClient(configuration *service.Configuration, httpClient *http.Client) *nghispersonserviceclient.APIClient {
	clientCfg := nghispersonserviceclient.Configuration{
		Host:          configuration.PersonClientHost,
		Scheme:        configuration.PersonClientScheme,
		DefaultHeader: map[string]string{},
		Debug:         false,
		UserAgent:     "",
		Servers: nghispersonserviceclient.ServerConfigurations{
			nghispersonserviceclient.ServerConfiguration{
				URL: configuration.PersonClientContext,
			},
		},
		OperationServers: map[string]nghispersonserviceclient.ServerConfigurations{},
		HTTPClient:       httpClient,
	}

	return nghispersonserviceclient.NewAPIClient(&clientCfg)
}

func spanNameFormatter(_ string, r *http.Request) string {
	name := r.Context().Value(appCtx.CLIENT_NAME)
	op := r.Context().Value(appCtx.CLIENT_OP)
	return fmt.Sprintf("%s: %s", name, op)
}
