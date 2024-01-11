package controllers

import (
	// import the package with the generated protobuf files
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	pb "github.com/wurt83ow/tinyurl/internal/controllers/proto"
	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/services/shorturl"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// UsersServer supports all necessary server methods.
type UsersServer struct {
	storage Storage
	options Options
	log     Log
	worker  Worker
	authz   Authz
	// need to embed type pb.Unimplemented<TypeName>
	// for compatibility with future versions
	pb.UnimplementedURLServiceServer
}

// NewUsersServer creates a new UsersServer instance.
func NewUsersServer(storage Storage, options Options, log Log, worker Worker, authz Authz) *UsersServer {

	instance := &UsersServer{
		storage: storage,
		options: options,
		log:     log,
		worker:  worker,
		authz:   authz,
		// need to embed type pb.Unimplemented<TypeName>
		// for compatibility with future versions
		UnimplementedURLServiceServer: pb.UnimplementedURLServiceServer{},
	}

	return instance
}

// ShortenURL implements the ShortenURL method from the Users protobuf service.
func (s *UsersServer) ShortenURL(ctx context.Context, req *pb.AddURLRequest) (*pb.AddURLResponse, error) {
	// Authenticate the user
	userID, err := s.authenticate(ctx)
	if err != nil {
		// Return an error instead of nil to inform the client about an authentication error
		return nil, err
	}

	// Describe the logic for processing a URL shortening request
	fullURL := req.GetFullurl()

	// Get the address for the short link from the settings
	shortURLAddress := s.options.ShortURLAdress()

	// Shorten the URL
	_, shortenedURL := shorturl.Shorten(fullURL, shortURLAddress)

	// Write the data to the database
	dataURL := models.DataURL{
		ShortURL:    shortenedURL,
		OriginalURL: fullURL,
		UserID:      userID,
	}

	// Use the InsertURL method from the repository
	_, err = s.storage.InsertURL(shortenedURL, dataURL)
	if err != nil {
		// Return an error to the client with an error code and an error message
		return nil, status.Errorf(codes.Internal, "failed to save URL to storage: %v", err)
	}

	// Return the response
	response := &pb.AddURLResponse{
		Shurl: shortenedURL,
	}

	// Display information about the user and shortened link
	fmt.Println("User ID:", userID)
	fmt.Println("Shortened URL:", shortenedURL)

	return response, nil
}

// getFullURL implements the getFullURL method from the Users protobuf service.
func (s *UsersServer) GetFullURL(ctx context.Context, req *pb.GetURLRequest) (*pb.GetURLResponse, error) {
	// Extract the key from the request
	key := req.GetKey()

	// Return the NotFound error code if the key is empty
	if key == "" {
		return nil, status.Error(codes.InvalidArgument, "empty key")
	}

	// Get the full URL from the storage
	data, err := s.storage.GetURL(key)

	// Return a NotFound error code if the URL was not found or an error occurred
	if err != nil || data.OriginalURL == "" {
		return nil, status.Error(codes.NotFound, "URL not found")
	}

	// Return the answer
	response := &pb.GetURLResponse{
		OriginalUrl: data.OriginalURL,
	}

	return response, nil
}

// DeleteUserURLs implements a method for deleting user URLs from the Users protobuf service.
func (s *UsersServer) DeleteUserURLs(ctx context.Context, req *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	// Get a list of URL IDs from the request
	ids := req.GetUrls()

	// Get the user ID from the context
	userID, err := s.authenticate(ctx)
	if err != nil {
		// Return an error instead of nil to inform the client about an authentication error
		return nil, err
	}

	// Add tasks to worker for asynchronous removal
	s.worker.Add(models.DeleteURL{UserID: userID, ShortURLs: ids})

	// Return the answer
	response := &pb.DeleteUserURLsResponse{
		Key: userID,
	}
	return response, nil
}

func (s *UsersServer) authenticate(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("failed to get metadata")
	}

	authHeader, ok := md["authorization"]
	if !ok || len(authHeader) == 0 {
		return "", errors.New("authorization header is missing")
	}

	// Describe the logic for checking the JWT token and retrieving the userID
	userID, err := s.authz.DecodeJWTToUser(authHeader[0])
	if err != nil {
		return "", err
	}

	return userID, nil
}

// RegisterUser handles user registration in gRPC.
func (s *UsersServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	// Extract the request parameters from the protobuf structure
	email := req.GetEmail()
	password := req.GetPassword()
	name := req.GetName()

	// Check if a user with this email exists
	_, err := s.storage.GetUser(email)
	if err == nil {
		return nil, status.Errorf(codes.AlreadyExists, "User with email %s already exists", email)
	}

	// Generate a password hash
	hash := s.authz.GetHash(email, password)

	// Create a DataUser object to save to storage
	dataUser := &models.DataUser{
		UUID:  uuid.New().String(),
		Email: email,
		Hash:  hash,
		Name:  name,
	}

	// Save the user to storage
	_, err = s.storage.InsertUser(email, *dataUser)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to register user: %v", err)
	}

	// Return a successful response
	return &pb.RegisterUserResponse{
		Message: fmt.Sprintf("User %s registered successfully", email),
	}, nil
}

// Login implements the Login method from the Users protobuf service.
func (s *UsersServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	// Extract data from the request
	email := req.GetEmail()
	password := req.GetPassword()

	// Get the user from the storage via email
	user, err := s.storage.GetUser(email)
	if err != nil {
		// If the user does not exist, then return an error
		return nil, status.Errorf(codes.NotFound, "User not found")
	}

	// Check the password
	if bytes.Equal(user.Hash, s.authz.GetHash(email, password)) {
		// Generate a new JWT token for the user
		freshToken := s.authz.CreateJWTTokenForUser(user.UUID)

		// Return a successful response with a token
		return &pb.LoginResponse{
			Token: freshToken,
		}, nil
	}

	// If the password is incorrect, return an authentication error
	return nil, status.Errorf(codes.Unauthenticated, "Incorrect email/password")
}
