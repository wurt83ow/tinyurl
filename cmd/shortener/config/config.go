package config

import (
	"flag"
	"os"
)

type Options struct {
	flagRunAddr        string
	flagShortURLAdress string
}

func NewOptions() *Options {
	return &Options{
		flagRunAddr:        "",
		flagShortURLAdress: "",
	}
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func (o *Options) ParseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&o.flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&o.flagShortURLAdress, "b", "http://localhost:8080/", "server`s address for shor url")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		o.flagRunAddr = envRunAddr
	}

	if envShortURLAdress := os.Getenv("BASE_URL"); envShortURLAdress != "" {
		o.flagShortURLAdress = envShortURLAdress
	}
}

func (o *Options) RunAddr() string {
	return o.flagRunAddr
}

func (o *Options) ShortURLAdress() string {
	return o.flagShortURLAdress
}
