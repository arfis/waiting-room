package client

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/arfis/waiting-room/nghis-adapter/internal/config/service"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewHTTPClient prepares the underlying HTTP client for the service clients
func NewHTTPClient(
	configuration *service.Configuration,
	logger *slog.Logger,
) *http.Client {
	return &http.Client{
		Transport: AuthorizationTransport(
			LoggerTransport(
				otelhttp.NewTransport(&http.Transport{}, otelhttp.WithSpanNameFormatter(spanNameFormatter)),
				logger,
			)),
		Timeout: time.Second * configuration.HTTPClientTimeout,
	}
}
