package logger

import (
	"git.prosoftke.sk/nghis/modules/go-libraries/logging"
	"github.com/arfis/waiting-room/nghis-adapter/internal/config/service"

	"log/slog"
	"os"
	"strings"
)

func NewLogging(cfg *service.Configuration) *slog.Logger {
	var contextFields []logging.ContextKey

	if len(cfg.ContextFields) > 0 {
		for _, i := range cfg.ContextFields {
			contextFields = append(contextFields, logging.ContextKey(strings.TrimSpace(i)))
		}
	}

	return logging.NewJSON(logging.Configuration{
		Level:         cfg.LogLevel,
		ContextFields: contextFields,
	}, os.Stdout)
}
