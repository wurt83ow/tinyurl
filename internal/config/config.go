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
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Options holds the configuration options for the application.
type Options struct {
	flagRunAddr         string
	flagShortURLAdress  string
	flagLogLevel        string
	flagFileStoragePath string
	flagDataBaseDSN     string
	flagJWTSigningKey   string
	flagConfigFile      string
	flagEnableHTTPS     bool
	flagHTTPSCertFile   string
	flagHTTPSKeyFile    string
	flagTrustedSubnet   string
}

// NewOptions creates a new instance of Options.
func NewOptions() *Options {
	return &Options{}
}

// ParseFlags parses the command line arguments and sets the corresponding option values.
func (o *Options) ParseFlags() {
	regStringVar(&o.flagRunAddr, "a", ":8080", "address and port to run server")
	regStringVar(&o.flagShortURLAdress, "b", "http://localhost:8080/", "server`s address for shor url")
	regStringVar(&o.flagLogLevel, "l", "info", "log level")
	regStringVar(&o.flagFileStoragePath, "f", "test777", "")
	regStringVar(&o.flagDataBaseDSN, "d", "", "")
	regStringVar(&o.flagJWTSigningKey, "j", "test_key", "jwt signing key")
	regStringVar(&o.flagConfigFile, "c", "", "path to configuration file in JSON format")
	regBoolVar(&o.flagEnableHTTPS, "s", false, "enable https")
	regStringVar(&o.flagHTTPSCertFile, "r", "", "path to https cert file")
	regStringVar(&o.flagHTTPSKeyFile, "k", "", "path to https key file")
	regStringVar(&o.flagTrustedSubnet, "t", "", "trusted subnet")
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

	if envHTTPSCertFile := os.Getenv("HTTPS_CERT_FILE"); envHTTPSCertFile != "" {
		o.flagHTTPSCertFile = envHTTPSCertFile
	}

	if envHTTPSKeyFile := os.Getenv("HTTPS_KEY_FILE"); envHTTPSKeyFile != "" {
		o.flagHTTPSKeyFile = envHTTPSKeyFile
	}

	if envTrustedSubnet := os.Getenv("TRUSTED_SUBNET"); envTrustedSubnet != "" {
		o.flagTrustedSubnet = envTrustedSubnet
	}

	if envConfigFile := os.Getenv("CONFIG"); envConfigFile != "" {
		o.flagConfigFile = envConfigFile
	}

	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS != "" {
		// Assuming "ENABLE_HTTPS" should be a boolean value
		enableHTTPS, err := strconv.ParseBool(envEnableHTTPS)
		if err == nil {
			o.flagEnableHTTPS = enableHTTPS
		} else {
			// Handle the error (failed to parse as boolean)
			fmt.Println("Failed to parse ENABLE_HTTPS as a boolean value:", err)
		}
	}

	// Check if config file path is provided. Redefine the parameters
	//if they are present in the file
	if o.flagConfigFile != "" {
		// Load options from config file
		err := o.LoadFromConfigFile(o.flagConfigFile)
		if err != nil {
			fmt.Println(o.flagConfigFile)
			fmt.Println("Error loading configuration from file:", err)
		}
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

// HTTPSCertFile returns the path to HTTPS cert file.
func (o *Options) HTTPSCertFile() string {
	return getStringFlag("r")
}

// HTTPSCertFile returns the path to HTTPS key file.
func (o *Options) HTTPSKeyFile() string {
	return getStringFlag("k")
}

// HTTPSCertFile returns the path to HTTPS key file.
func (o *Options) TrustedSubnet() string {
	return getStringFlag("t")
}

// EnableHTTPS returns whether HTTPS is enabled.
func (o *Options) EnableHTTPS() bool {
	return getBoolFlag("s")
}

// regStringVar registers a string flag with the specified name, default value, and usage string.
func regStringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

// regBoolVar registers a bool flag with the specified name, default value, and usage string.
func regBoolVar(p *bool, name string, value bool, usage string) {
	if flag.Lookup(name) == nil {
		flag.BoolVar(p, name, value, usage)
	}
}

// getStringFlag retrieves the string value of the specified flag.
func getStringFlag(name string) string {
	return flag.Lookup(name).Value.(flag.Getter).Get().(string)
}

// getBoolFlag retrieves the bool value of the specified flag.
func getBoolFlag(name string) bool {
	return flag.Lookup(name).Value.(flag.Getter).Get().(bool)
}

// GetAsString reads an environment variable or returns a default value.
func GetAsString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}

// loadFromConfigFile loads configuration options from a JSON file.
func (o *Options) LoadFromConfigFile(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := make(map[string]interface{})
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}

	// Set values from config file
	o.setIfNotEmpty(&o.flagRunAddr, config["server_address"])
	o.setIfNotEmpty(&o.flagShortURLAdress, config["base_url"])
	o.setIfNotEmpty(&o.flagLogLevel, config["log_level"])
	o.setIfNotEmpty(&o.flagFileStoragePath, config["file_storage_path"])
	o.setIfNotEmpty(&o.flagDataBaseDSN, config["database_dsn"])
	o.setIfNotEmpty(&o.flagJWTSigningKey, config["jwt_signing_key"])
	o.setIfNotEmpty(&o.flagHTTPSCertFile, config["https_cert_file"])
	o.setIfNotEmpty(&o.flagHTTPSKeyFile, config["https_key_file"])
	o.setIfNotEmpty(&o.flagTrustedSubnet, config["trusted_subnet"])

	// Handle boolean value for enable_https
	if enableHTTPS, ok := config["enable_https"].(bool); ok {
		o.flagEnableHTTPS = enableHTTPS
	}

	return nil
}

// setIfNotEmpty sets the target variable if the value is not empty.
func (o *Options) setIfNotEmpty(target *string, value interface{}) {
	if *target != "" {
		// If the value is already set, return from the function
		return
	}
	if strValue, ok := value.(string); ok && strValue != "" {
		*target = strValue
	}
}
