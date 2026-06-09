package core_worker

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type ExecutorConfig struct {
	HTTPAllowlist []string `envconfig:"ALLOWLIST"`
}

func NewConfig() (ExecutorConfig, error) {
	var config ExecutorConfig

	if err := envconfig.Process("EXECUTOR", &config); err != nil {
		return ExecutorConfig{}, fmt.Errorf("process envconfig: %w", err)
	}

	return config, nil
}

func NewConfigMust() ExecutorConfig { // rework
	config, err := NewConfig()
	if err != nil {
		panic(err)
	}

	return config
}
