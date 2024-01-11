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
	"github.com/wurt83ow/tinyurl/internal/storage"
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

// ShortenBatch implements the ShortenBatch method from the URLService protobuf service.
func (s *UsersServer) ShortenBatch(ctx context.Context, req *pb.ShortenBatchRequest) (*pb.ShortenBatchResponse, error) {
	// Get the user ID from the context
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	// Get the short URL address from the options
	shortURLAdress := s.options.ShortURLAdress()

	// Initialize a map to store data about shortened URLs
	dataURL := make(storage.StorageURL)

	// Initialize the response for the client
	resp := &pb.ShortenBatchResponse{}

	// Iterate through each URL in the batch
	for _, url := range req.Urls {
		// Shorten the original URL
		key, shurl := shorturl.Shorten(url.OriginalUrl, shortURLAdress)

		// Save the full URL to storage with the key received earlier
		data := models.DataURL{UUID: url.Uuid, ShortURL: shurl, OriginalURL: url.OriginalUrl, UserID: userID}
		dataURL[key] = data

		// Add the shortened URL to the response
		resp.Urls = append(resp.Urls, &pb.ShortenedURL{
			Uuid:     url.Uuid,
			ShortUrl: shurl,
		})
	}

	// Insert the batch of URLs into storage
	err = s.storage.InsertBatch(dataURL)
	if err != nil {
		return nil, status.Error(codes.Internal, "Error inserting batch into storage")
	}

	// Return a successful response
	return resp, nil
}

// ShortenJSON is a handler method for shortening a single URL from a JSON request.
// It takes a context.Context, *pb.ShortenJSONRequest, and returns *pb.ShortenJSONResponse and error.
func (s *UsersServer) ShortenJSON(ctx context.Context, req *pb.ShortenJSONRequest) (*pb.ShortenJSONResponse, error) {
	// Convert request model from protobuf to internal model
	internalReq := &models.Request{
		URL: req.GetUrl(),
	}

	// Deserialize the request into the model structure
	s.log.Info("decoding request")

	// Check if the request JSON body is empty
	if internalReq.URL == "" {
		// Log the error and respond with a Bad Request status code
		s.log.Info("request JSON body is empty")
		return nil, status.Error(codes.InvalidArgument, "request JSON body is empty")
	}

	// Get the short URL address from the options
	shortURLAdress := s.options.ShortURLAdress()

	// Shorten the original URL
	key, shurl := shorturl.Shorten(internalReq.URL, shortURLAdress)

	// Retrieve the user ID from the request context
	userID, err := s.authenticate(ctx)
	if err != nil {

		// Return an error instead of nil to inform the client about an authentication error
		return nil, err
	}

	// Save the full URL to storage with the key received earlier
	m, err := s.storage.InsertURL(key, models.DataURL{ShortURL: shurl, OriginalURL: internalReq.URL, UserID: userID})
	if err != nil {
		if err == storage.ErrConflict {
			// Respond with a Conflict status code for conflicts
			return nil, status.Error(codes.AlreadyExists, "URL conflict")
		} else {
			// Respond with a Bad Request status code for other errors
			return nil, status.Error(codes.Internal, "error inserting URL into storage")
		}
	}

	// Convert response model from internal to protobuf
	internalResp := &models.Response{
		Result: m.ShortURL,
	}

	// Fill in the response model
	resp := &pb.ShortenJSONResponse{
		Result: internalResp.Result,
	}

	// Log the successful response
	s.log.Info("sending response")

	return resp, nil
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

// GetUserURLs retrieves user URLs.
func (s *UsersServer) GetUserURLs(ctx context.Context, req *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	// Get the user ID from the context
	userID, err := s.authenticate(ctx)
	if err != nil {
		return nil, err
	}

	// Get URLs associated with the user from the storage
	data := s.storage.GetUserURLs(userID)

	// Convert data to the format expected by the client
	var userURLs []*pb.UserURL
	for _, url := range data {
		userURL := &pb.UserURL{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		}
		userURLs = append(userURLs, userURL)
	}

	// Formulate the response
	response := &pb.GetUserURLsResponse{
		Urls: userURLs,
	}

	return response, nil
}

// HealthCheck checks storage availability and returns the appropriate status.
func (s *UsersServer) HealthCheck(ctx context.Context, req *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	if s.storage.GetBaseConnection() {
		// Storage available
		return &pb.HealthCheckResponse{
			Status: pb.HealthCheckResponse_OK,
		}, nil
	}

	// Storage unavailable
	return &pb.HealthCheckResponse{
		Status: pb.HealthCheckResponse_ERROR,
	}, nil
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
