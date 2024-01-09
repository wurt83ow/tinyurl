package controllers

import (
	// импортируем пакет со сгенерированными protobuf-файлами
	"context"
	"fmt"

	pb "github.com/wurt83ow/tinyurl/internal/controllers/proto"
)

// UsersServer поддерживает все необходимые методы сервера.
type UsersServer struct {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	pb.UnimplementedUsersServer
}

// NewUsersServer создает новый экземпляр UsersServer.
func NewUsersServer() *UsersServer {
	// нужно встраивать тип pb.Unimplemented<TypeName>
	// для совместимости с будущими версиями
	return &UsersServer{
		UnimplementedUsersServer: pb.UnimplementedUsersServer{},
	}
}

// ShortenURL реализует метод ShortenURL из protobuf-сервиса Users.
func (s *UsersServer) ShortenURL(ctx context.Context, req *pb.AddURLRequest) (*pb.AddURLResponse, error) {
	// Ваша логика обработки запроса на сокращение URL
	fullURL := req.GetFullurl()

	// Здесь вы можете добавить логику для сокращения URL
	// Например, можно использовать внешнюю библиотеку для генерации коротких URL
	shortURL := generateShortURL(fullURL)

	// Возврат ответа
	response := &pb.AddURLResponse{
		Shurl: shortURL,
	}
	fmt.Println("7777777777777777777777777777777777777", response)

	return response, nil
}

// Пример функции для генерации короткого URL (замените на свою логику)
func generateShortURL(fullURL string) string {
	// Ваша логика для генерации короткого URL
	// Например, можно использовать хэш-функции или другие методы
	return fmt.Sprintf("http://short.url/%s", fullURL)
}
