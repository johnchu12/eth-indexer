// environment/environment.go

package environment

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

// LoadConfig loads the environment configuration for the specified service.
func LoadConfig(serviceName string, cfg interface{}) error {
	// Get the current working directory.
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Construct the .env file path (always under the env folder at the project root).
	envPath := filepath.Join(currentDir, "env", fmt.Sprintf("%s.env", serviceName))

	// Check if the .env file exists.
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return fmt.Errorf("env file not found for service %s: %s", serviceName, envPath)
	}

	// Load the .env file.
	if err := godotenv.Load(envPath); err != nil {
		return fmt.Errorf("failed to load env file (%s): %w", envPath, err)
	}

	log.Printf("Successfully loaded environment config for %s: %s", serviceName, envPath)

	// Map environment variables to the config struct.
	if err := envconfig.Process("", cfg); err != nil {
		return fmt.Errorf("failed to process environment variables: %w", err)
	}

	// Check if the config is empty.
	if reflect.DeepEqual(cfg, reflect.Zero(reflect.TypeOf(cfg)).Interface()) {
		return fmt.Errorf("configuration is empty: %+v", cfg)
	}

	return nil
}

// ListEnvironmentVariables returns a formatted string of all environment variables.
func ListEnvironmentVariables() string {
	// Get all environment variables.
	envVars := os.Environ()

	// Sort the environment variables.
	sort.Strings(envVars)

	// Build the formatted string.
	var builder strings.Builder
	builder.WriteString("Environment Variables:\n")
	for _, envVar := range envVars {
		builder.WriteString(fmt.Sprintf("  %s\n", envVar))
	}

	return builder.String()
}
