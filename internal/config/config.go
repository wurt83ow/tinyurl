// Package config provides a configuration management mechanism for the application.
// It includes a set of options that can be configured through command line arguments
// and environment variables.
//
// The Options struct holds the configuration options, and NewOptions creates a new
// instance of Options with default values. The ParseFlags method parses command line
// arguments and sets corresponding option values. The package also provides getter
// methods for accessing specific configuration parameters.
//
// Usage Example:
//
//	// Create a new instance of Options
//	options := config.NewOptions()
//
//	// Parse command line arguments and set option values
//	options.ParseFlags()
//
//	// Access specific configuration parameters
//	runAddr := options.RunAddr()
//	shortURLAddr := options.ShortURLAdress()
//	logLevel := options.LogLevel()
//	fileStoragePath := options.FileStoragePath()
//	dataBaseDSN := options.DataBaseDSN()
//	jwtSigningKey := options.JWTSigningKey()
//
//	// Alternatively, configure options through environment variables
//	os.Setenv("SERVER_ADDRESS", ":8081")
//	os.Setenv("BASE_URL", "http://localhost:8081/")
//	os.Setenv("LOG_LEVEL", "debug")
//
//	// Re-create options and parse environment variables
//	options = config.NewOptions()
//	options.ParseFlags()
//
//	// Access specific configuration parameters again
//	updatedRunAddr := options.RunAddr()
//	updatedShortURLAddr := options.ShortURLAdress()
//	updatedLogLevel := options.LogLevel()
package config

import (
	"flag"
	"os"
)

// Options holds the configuration options for the application.
type Options struct {
	flagRunAddr         string
	flagShortURLAdress  string
	flagLogLevel        string
	flagFileStoragePath string
	flagDataBaseDSN     string
	flagJWTSigningKey   string
}

// NewOptions creates a new instance of Options.
func NewOptions() *Options {
	return &Options{}
}

// ParseFlags parses the command line arguments and sets the corresponding option values.
func (o *Options) ParseFlags() {
	regStringVar(&o.flagRunAddr, "a", ":8080", "address and port to run server")
	regStringVar(&o.flagShortURLAdress, "b", "http://localhost:8080/", "server's address for short URL")
	regStringVar(&o.flagLogLevel, "l", "info", "log level")
	regStringVar(&o.flagFileStoragePath, "f", "/tmp/short-url-db.json", "file storage path")
	regStringVar(&o.flagDataBaseDSN, "d", "", "database DSN")
	regStringVar(&o.flagJWTSigningKey, "j", "test_key", "JWT signing key")

	// parse the arguments passed to the server into registered variables
	flag.Parse()

	// Check if corresponding environment variables are set and override the values if present.
	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		o.flagRunAddr = envRunAddr
	}

	if envShortURLAdress := os.Getenv("BASE_URL"); envShortURLAdress != "" {
		o.flagShortURLAdress = envShortURLAdress
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		o.flagLogLevel = envLogLevel
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		o.flagFileStoragePath = envFileStoragePath
	}

	if envDataBaseDSN := os.Getenv("DATABASE_DSN"); envDataBaseDSN != "" {
		o.flagDataBaseDSN = envDataBaseDSN
	}

	if envJWTSigningKey := os.Getenv("JWT_SIGNING_KEY"); envJWTSigningKey != "" {
		o.flagJWTSigningKey = envJWTSigningKey
	}
}

// RunAddr returns the configured server address.
func (o *Options) RunAddr() string {
	return getStringFlag("a")
}

// ShortURLAdress returns the configured server's address for short URLs.
func (o *Options) ShortURLAdress() string {
	return getStringFlag("b")
}

// LogLevel returns the configured log level.
func (o *Options) LogLevel() string {
	return getStringFlag("l")
}

// FileStoragePath returns the configured file storage path.
func (o *Options) FileStoragePath() string {
	return getStringFlag("f")
}

// DataBaseDSN returns the configured database DSN.
func (o *Options) DataBaseDSN() string {
	return getStringFlag("d")
}

// JWTSigningKey returns the configured JWT signing key.
func (o *Options) JWTSigningKey() string {
	return getStringFlag("j")
}

// regStringVar registers a string flag with the specified name, default value, and usage string.
func regStringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

// getStringFlag retrieves the string value of the specified flag.
func getStringFlag(name string) string {
	return flag.Lookup(name).Value.(flag.Getter).Get().(string)
}

// GetAsString reads an environment variable or returns a default value.
func GetAsString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
