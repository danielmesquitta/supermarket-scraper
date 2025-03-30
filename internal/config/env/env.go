package env

import (
	"bytes"
	_ "embed" // embedding env files
	"fmt"
	"log"
	"os"

	root "github.com/danielmesquitta/supermarket-web-scraper"
	"github.com/danielmesquitta/supermarket-web-scraper/internal/pkg/validator"
	"github.com/spf13/viper"
)

const defaultEnvFileName = ".env"

type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentProduction  Environment = "production"
	EnvironmentStaging     Environment = "staging"
	EnvironmentTest        Environment = "test"
)

type Env struct {
	v validator.Validator

	Environment Environment `mapstructure:"ENVIRONMENT" validate:"required,oneof=development production staging test"`
	ChromePath  string      `mapstructure:"CHROME_PATH" validate:"required"`
	CDPPort     int         `mapstructure:"CDP_PORT"    validate:"required"`
}

func NewEnv(v validator.Validator) *Env {
	e := &Env{
		v: v,
	}

	if err := e.loadEnv(); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}

	return e
}

func (e *Env) loadEnv() error {
	envFile, err := e.getEnvFile()
	if err != nil {
		return fmt.Errorf("failed to get environment file: %w", err)
	}

	viper.SetConfigType("env")

	if err := viper.ReadConfig(bytes.NewBuffer(envFile)); err != nil {
		return fmt.Errorf("failed to read environment file: %w", err)
	}

	viper.AutomaticEnv()

	if err := viper.Unmarshal(&e); err != nil {
		return fmt.Errorf("failed to unmarshal environment file: %w", err)
	}

	if err := e.validate(); err != nil {
		return fmt.Errorf("failed to validate environment file: %w", err)
	}

	return nil
}

func (e *Env) getEnvFile() (envFile []byte, err error) {
	environment := os.Getenv("ENVIRONMENT")

	if environment != "" {
		envFileName := fmt.Sprintf("%s.%s", defaultEnvFileName, environment)
		envFile, err = root.Env.ReadFile(envFileName)
		if err == nil {
			return envFile, nil
		}
	}

	envFile, err = root.Env.ReadFile(defaultEnvFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file: %w", err)
	}

	return envFile, nil
}

func (e *Env) validate() error {
	if err := e.v.Validate(e); err != nil {
		return fmt.Errorf("failed to validate environment file: %w", err)
	}
	if e.Environment == "" {
		e.Environment = EnvironmentDevelopment
	}

	return nil
}
