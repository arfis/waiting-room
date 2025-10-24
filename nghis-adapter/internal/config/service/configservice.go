package service

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Configuration struct {
	// service name
	ServiceName         string `env:"SERVICE_NAME,omitempty" env-default:"nghis-adapter"`
	ServiceInstanceName string `env:"SERVICE_INSTANCE_NAME,omitempty" env-default:"nghis-adapter-1"`

	// logging
	LogLevel      string   `env:"LOG_LEVEL,omitempty" env-default:"DEBUG"`
	ContextFields []string `env:"CONTEXT_FIELDS,omitempty"`

	// HTTP client
	HTTPClientTimeout time.Duration `env:"HTTP_CLIENT_TIMEOUT" env-default:"2s"`

	// clinical client
	ClinicalClientHost    string `env:"CLINICAL_CLIENT_HOST" env-default:"api.dev.nghis.prosoftke.sk"`
	ClinicalClientScheme  string `env:"CLINICAL_CLIENT_SCHEME" env-default:"https"`
	ClinicalClientContext string `env:"CLINICAL_CLIENT_CONTEXT" env-default:"clinical"`

	// person client
	PersonClientHost    string `env:"PERSON_CLIENT_HOST" env-default:"api.dev.nghis.prosoftke.sk"`
	PersonClientScheme  string `env:"PERSON_CLIENT_SCHEME" env-default:"https"`
	PersonClientContext string `env:"PERSON_CLIENT_CONTEXT" env-default:"person"`

	// server
	ServerPort    string        `env:"APP_PORT" env-default:"8060"`
	ServerContext string        `env:"APP_CONTEXT" env-default:"/nghis-adapter"`
	HTTPTimeout   time.Duration `env:"HTTP_TIMEOUT" env-default:"60s"`
}

func NewConfiguration() (*Configuration, error) {
	var configuration Configuration

	err := cleanenv.ReadEnv(&configuration)
	if err != nil {
		return nil, err
	}

	return &configuration, nil
}
