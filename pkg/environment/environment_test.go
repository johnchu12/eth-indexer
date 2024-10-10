package environment

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestConfig holds the test configuration.
type TestConfig struct {
	TestVar string `envconfig:"TEST_VAR"`
}

// TestLoadConfig tests the LoadConfig function.
func TestLoadConfig(t *testing.T) {
	// Create a temporary directory to simulate project structure
	tempDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create env directory in the temporary directory
	envDir := filepath.Join(tempDir, "env")
	err = os.Mkdir(envDir, 0o755)
	assert.NoError(t, err)

	// Create test .env file
	envContent := []byte("TEST_VAR=test_value")
	err = os.WriteFile(filepath.Join(envDir, "test_service.env"), envContent, 0o644)
	assert.NoError(t, err)

	// Change to temporary directory
	originalWd, _ := os.Getwd()
	err = os.Chdir(tempDir)
	assert.NoError(t, err)
	defer os.Chdir(originalWd)

	// Execute tests
	var cfg TestConfig
	err = LoadConfig("test_service", &cfg)
	assert.NoError(t, err)
	assert.Equal(t, "test_value", cfg.TestVar)

	// Test error case: non-existent service
	err = LoadConfig("non_existent_service", &cfg)
	assert.Error(t, err)
}

// TestListEnvironmentVariables tests the ListEnvironmentVariables function.
func TestListEnvironmentVariables(t *testing.T) {
	// Set some test environment variables
	os.Setenv("TEST_VAR1", "value1")
	os.Setenv("TEST_VAR2", "value2")
	defer os.Unsetenv("TEST_VAR1")
	defer os.Unsetenv("TEST_VAR2")

	// Get the list of environment variables
	list := ListEnvironmentVariables()

	// Verify the results
	assert.Contains(t, list, "Environment Variables:")
	assert.Contains(t, list, "TEST_VAR1=value1")
	assert.Contains(t, list, "TEST_VAR2=value2")
}
