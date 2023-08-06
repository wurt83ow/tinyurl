package config

import (
	"flag"
)

var options struct {
	flagRunAddr        string
	flagShortURLAdress string
}

// parseFlags обрабатывает аргументы командной строки
// и сохраняет их значения в соответствующих переменных
func ParseFlags() {
	// регистрируем переменную flagRunAddr
	// как аргумент -a со значением :8080 по умолчанию
	flag.StringVar(&options.flagRunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&options.flagShortURLAdress, "b", "http://localhost:8080/", "server`s address for shor url")
	// парсим переданные серверу аргументы в зарегистрированные переменные
	flag.Parse()
}

func RunAddr() string {
	return options.flagRunAddr
}

func ShortURLAdress() string {
	return options.flagShortURLAdress
}
