package core_scheduler

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	IsTestMode bool `envconfig:"IS_TEST_MODE" default:"false"`
}

func NewConfig() (Config, error) {
	var config Config

	if err := envconfig.Process("SCHEDULER", &config); err != nil {
		return Config{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() Config {
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}

	return config
}
