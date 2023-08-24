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
}

func NewOptions() *Options {
	return &Options{
		flagRunAddr:         "",
		flagShortURLAdress:  "",
		flagLogLevel:        "",
		flagFileStoragePath: "",
	}
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func (o *Options) ParseFlags() {

	regStringVar(&o.flagRunAddr, "a", ":8080", "address and port to run server")
	regStringVar(&o.flagShortURLAdress, "b", "http://localhost:8080/", "server`s address for shor url")
	regStringVar(&o.flagLogLevel, "l", "info", "log level")
	regStringVar(&o.flagFileStoragePath, "f", "/tmp/short-url-db.json", "default file storage path")

	// парсим переданные серверу аргументы в зарегистрированные переменные
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

func regStringVar(p *string, name string, value string, usage string) {
	if flag.Lookup(name) == nil {
		flag.StringVar(p, name, value, usage)
	}
}

func getStringFlag(name string) string {
	return flag.Lookup(name).Value.(flag.Getter).Get().(string)
}
