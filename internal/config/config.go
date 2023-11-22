package config

import (
	"flag"
	"os"
)

type Options struct {
	flagRunAddr         string
	flagShortURLAdress  string
	flagLogLevel        string
	flagFileStoragePath string
	flagDataBaseDSN     string
	flagJWTSigningKey   string
}

func NewOptions() *Options {
	return &Options{}
}

// parseFlags handles command line arguments
// and stores their values in the corresponding variables
func (o *Options) ParseFlags() {
	regStringVar(&o.flagRunAddr, "a", ":8080", "address and port to run server")
	regStringVar(&o.flagShortURLAdress, "b", "http://localhost:8080/", "server`s address for shor url")
	regStringVar(&o.flagLogLevel, "l", "info", "log level")
	regStringVar(&o.flagFileStoragePath, "f", "/tmp/short-url-db.json", "")
	regStringVar(&o.flagDataBaseDSN, "d", "", "")
	regStringVar(&o.flagJWTSigningKey, "j", "test_key", "jwt signing key")

	// parse the arguments passed to the server into registered variables
	flag.Parse()

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

func (o *Options) RunAddr() string {
	return getStringFlag("a")
}

func (o *Options) ShortURLAdress() string {
	return getStringFlag("b")
}

func (o *Options) LogLevel() string {
	return getStringFlag("l")
}

func (o *Options) FileStoragePath() string {
	return getStringFlag("f")
}

func (o *Options) DataBaseDSN() string {
	return getStringFlag("d")
}

func (o *Options) JWTSigningKey() string {
	return getStringFlag("j")
}

func regStringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

func getStringFlag(name string) string {
	return flag.Lookup(name).Value.(flag.Getter).Get().(string)
}

// GetAsString reads an environment or returns a default value
func GetAsString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return defaultValue
}
