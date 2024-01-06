package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShorten(t *testing.T) {

	option := NewOptions()
	option.ParseFlags()

	k, v := "testEnvKey", "testEnvValue"

	os.Setenv(k, v)
	t.Run("test func GetAsString", func(t *testing.T) {
		assert.Equal(t, GetAsString(k, ""), v, "the returned result does not match the expected one")
	})

	randomKey := "047012de-7ea9-417d-9799-6c4a5ed8df7a"
	t.Run("test func GetAsString default", func(t *testing.T) {
		assert.Equal(t, GetAsString(randomKey, "defaultValue"), "defaultValue", "the returned result does not match the expected one")
	})

	testCases := []struct {
		name     string
		testfunc func() string
		result   string
	}{
		{name: "test func RunAddr", testfunc: option.RunAddr, result: ":8080"},
		{name: "test func LogLevel", testfunc: option.LogLevel, result: "info"},
		{name: "test func FileStoragePath", testfunc: option.FileStoragePath, result: "test777"},
		{name: "test func JWTSigningKey", testfunc: option.JWTSigningKey, result: "test_key"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.testfunc(), tc.result, "the returned result does not match the expected one")
		})
	}
}

func TestOptions_ParseFlags(t *testing.T) {
	// Backup original command line arguments and restore them after the test
	originalArgs := os.Args
	defer func() { os.Args = originalArgs }()

	// Prepare test command line arguments
	testArgs := []string{
		"app", "-a", ":8081", "-b", "http://localhost:8081/", "-l", "info",
		"-f", "test777", "-d", "testdb", "-j", "test_key", "-c", "/path/to/config.json", "-s",
	}
	os.Args = testArgs

	// Prepare test environment variables
	os.Setenv("SERVER_ADDRESS", ":8081")
	os.Setenv("BASE_URL", "http://localhost:8081/")
	os.Setenv("LOG_LEVEL", "info")
	os.Setenv("FILE_STORAGE_PATH", "test888")
	os.Setenv("DATABASE_DSN", "testdb_env")
	os.Setenv("JWT_SIGNING_KEY", "test_key_env")
	os.Setenv("CONFIG", "/path/to/config.json")
	os.Setenv("ENABLE_HTTPS", "true")

	// Create an instance of Options
	options := NewOptions()

	// Parse command line arguments
	options.ParseFlags()

	// Check if the options are correctly set
	assert.Equal(t, ":8081", options.RunAddr())
	assert.Equal(t, "http://localhost:8081/", options.ShortURLAdress())
	assert.Equal(t, "info", options.LogLevel())
	assert.Equal(t, "test777", options.FileStoragePath())
	assert.Equal(t, "testdb", options.DataBaseDSN())
	assert.Equal(t, "test_key", options.JWTSigningKey())
	assert.Equal(t, "/path/to/config.json", options.flagConfigFile)
	assert.True(t, options.EnableHTTPS())

	// Reset the environment variables
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("JWT_SIGNING_KEY")
	os.Unsetenv("CONFIG")
	os.Unsetenv("ENABLE_HTTPS")

	// Check if the options are correctly set from environment variables
	assert.Equal(t, ":8081", options.RunAddr())
	assert.Equal(t, "http://localhost:8081/", options.ShortURLAdress())
	assert.Equal(t, "info", options.LogLevel())
	assert.Equal(t, "test777", options.FileStoragePath())
	assert.Equal(t, "testdb", options.DataBaseDSN())
	assert.Equal(t, "test_key", options.JWTSigningKey())
	assert.Equal(t, "/path/to/config.json", options.flagConfigFile)
	assert.True(t, options.EnableHTTPS())
}
