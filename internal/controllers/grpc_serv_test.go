package controllers_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	authz "github.com/wurt83ow/tinyurl/internal/authorization"
	"github.com/wurt83ow/tinyurl/internal/controllers"
	pb "github.com/wurt83ow/tinyurl/internal/controllers/proto"
	"github.com/wurt83ow/tinyurl/internal/models"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/metadata"
)

type mockOptions struct {
	parseFlagsFunc      func()
	runAddrFunc         func() string
	trustedSubnetFunc   func() string
	shortURLAddressFunc func() string
}

func (m *mockOptions) ParseFlags() {
	m.parseFlagsFunc()
}

func (m *mockOptions) RunAddr() string {
	return m.runAddrFunc()
}

func (m *mockOptions) TrustedSubnet() string {
	return m.trustedSubnetFunc()
}

func (m *mockOptions) ShortURLAdress() string {
	return m.shortURLAddressFunc()
}

type mockLog struct {
	infoFunc func(msg string, fields ...zapcore.Field)
}

func (m *mockLog) Info(msg string, fields ...zapcore.Field) {
	m.infoFunc(msg, fields...)
}

type mockWorker struct {
	addFunc func(task models.DeleteURL)
}

func (m *mockWorker) Add(task models.DeleteURL) {
	if m.addFunc != nil {
		m.addFunc(task)
	}
}

type mockAuthz struct {
	authCookieFunc         func(jwtToken, userID string) *http.Cookie
	createJWTTokenFunc     func(userID string) string
	getHashFunc            func(email, password string) []byte
	jwtAuthzMiddlewareFunc func(storage authz.Storage, log authz.Log) func(http.Handler) http.Handler
}

func (m *mockAuthz) AuthCookie(jwtToken, userID string) *http.Cookie {
	return m.authCookieFunc(jwtToken, userID)
}

func (m *mockAuthz) CreateJWTTokenForUser(userID string) string {
	return m.createJWTTokenFunc(userID)
}

func (m *mockAuthz) GetHash(email, password string) []byte {
	return m.getHashFunc(email, password)
}

func (m *mockAuthz) DecodeJWTToUser(token string) (string, error) {
	// Mock implementation
	return "mockUserID", nil
}

func (m *mockAuthz) JWTAuthzMiddleware(storage authz.Storage, log authz.Log) func(http.Handler) http.Handler {
	return m.jwtAuthzMiddlewareFunc(storage, log)
}

type mockStorage struct {
	insertURLFunc            func(string, models.DataURL) (models.DataURL, error)
	insertBatchFunc          func(map[string]models.DataURL) error
	getURLFunc               func(string) (models.DataURL, error)
	getUserURLsFunc          func(string) []models.DataURLite
	deleteUserURLsFunc       func(string, []string)
	deleteURLsFunc           func(...models.DeleteURL) error
	getBaseConnectionFunc    func() bool
	getUserFunc              func(string) (models.DataUser, error)
	getUsersAndURLsCountFunc func() (int, int, error)
	insertUserFunc           func(string, models.DataUser) (models.DataUser, error)
	saveBatchFunc            func(map[string]models.DataURL) error
	saveURLFunc              func(string, models.DataURL) (models.DataURL, error)
	saveUserFunc             func(string, models.DataUser) (models.DataUser, error)
}

func (m *mockStorage) InsertURL(key string, data models.DataURL) (models.DataURL, error) {
	return m.insertURLFunc(key, data)
}

func (m *mockStorage) InsertBatch(data map[string]models.DataURL) error {
	return m.insertBatchFunc(data)
}

func (m *mockStorage) GetURL(key string) (models.DataURL, error) {
	return m.getURLFunc(key)
}

func (m *mockStorage) GetUserURLs(userID string) []models.DataURLite {
	return m.getUserURLsFunc(userID)
}

func (m *mockStorage) DeleteUserURLs(userID string, shortURLs []string) {
	m.deleteUserURLsFunc(userID, shortURLs)
}

func (m *mockStorage) DeleteURLs(urls ...models.DeleteURL) error {
	return m.deleteURLsFunc(urls...)
}

func (m *mockStorage) GetBaseConnection() bool {
	return m.getBaseConnectionFunc()
}

func (m *mockStorage) GetUser(email string) (models.DataUser, error) {
	return m.getUserFunc(email)
}

func (m *mockStorage) GetUsersAndURLsCount() (int, int, error) {
	return m.getUsersAndURLsCountFunc()
}

func (m *mockStorage) InsertUser(email string, data models.DataUser) (models.DataUser, error) {
	return m.insertUserFunc(email, data)
}

func (m *mockStorage) SaveBatch(data map[string]models.DataURL) error {
	return m.saveBatchFunc(data)
}

func (m *mockStorage) SaveURL(key string, data models.DataURL) (models.DataURL, error) {
	return m.saveURLFunc(key, data)
}

func (m *mockStorage) SaveUser(email string, data models.DataUser) (models.DataUser, error) {
	return m.saveUserFunc(email, data)
}

// TestContext contains dependencies for test functions.
type TestContext struct {
	t           *testing.T
	FakeToken   string
	Server      *controllers.UsersServer
	MockStorage *mockStorage
	MockOptions *mockOptions
	MockWorker  *mockWorker
	MockAuthz   *mockAuthz
}

// NewTestContext creates a new testing context.
func NewTestContext(t *testing.T) *TestContext {
	mockStorage := &mockStorage{
		insertURLFunc: func(key string, data models.DataURL) (models.DataURL, error) {
			return data, nil
		},
		deleteURLsFunc: func(urls ...models.DeleteURL) error {
			// Code for checking and/or emulating method behavior DeleteURLs
			return nil
		},
		getBaseConnectionFunc: func() bool {
			// Code for checking and/or emulating method behavior GetBaseConnection
			return true
		},
		getUsersAndURLsCountFunc: func() (int, int, error) {
			// Code for checking and/or emulating method behavior GetUsersAndURLsCount
			return 0, 0, nil
		},
		insertUserFunc: func(email string, data models.DataUser) (models.DataUser, error) {
			// Code for checking and/or emulating method behavior InsertUser
			return data, nil
		},
		saveBatchFunc: func(data map[string]models.DataURL) error {
			// Code for checking and/or emulating method behavior SaveBatch
			return nil
		},
		saveURLFunc: func(key string, data models.DataURL) (models.DataURL, error) {
			// Code for checking and/or emulating method behavior SaveURL
			return data, nil
		},
		saveUserFunc: func(email string, data models.DataUser) (models.DataUser, error) {
			// Code for checking and/or emulating method behavior SaveUser
			return data, nil
		},
	}

	mockOptions := &mockOptions{
		parseFlagsFunc: func() {
			// Code for checking and/or emulating method behavior ParseFlags
		},
		runAddrFunc: func() string {
			// Code for checking and/or emulating method behavior RunAddr
			return "mockedAddress"
		},
		trustedSubnetFunc: func() string {
			// Code for checking and/or emulating method behavior TrustedSubnet
			return "mockedSubnet"
		},
		shortURLAddressFunc: func() string {
			// Code for checking and/or emulating method behavior ShortURLAdress
			return "mockedShortURLAddress"
		},
	}
	mockLog := &mockLog{
		infoFunc: func(msg string, fields ...zapcore.Field) {
			// Code for checking and/or emulating method behavior Info
		},
	}
	mockWorker := &mockWorker{}

	fakeToken := "mockedJWTToken"

	mockAuthz := &mockAuthz{
		authCookieFunc: func(jwtToken, userID string) *http.Cookie {
			assert.Equal(t, fakeToken, jwtToken) // Check that the token was transferred correctly
			return &http.Cookie{Name: "mockedCookie", Value: "mockedValue"}
		},
		createJWTTokenFunc: func(userID string) string {
			// Code for checking and/or emulating method behavior CreateJWTTokenForUser
			return "mockedJWTToken"
		},
		getHashFunc: func(email, password string) []byte {
			// Code for checking and/or emulating method behavior GetHash
			return []byte("mockedHash")
		},
		jwtAuthzMiddlewareFunc: func(storage authz.Storage, log authz.Log) func(http.Handler) http.Handler {
			// Code for checking and/or emulating method behavior JWTAuthzMiddleware
			return func(next http.Handler) http.Handler {
				return next
			}
		},
	}

	server := controllers.NewUsersServer(mockStorage, mockOptions, mockLog, mockWorker, mockAuthz)

	return &TestContext{
		t:           t,
		FakeToken:   fakeToken,
		Server:      server,
		MockStorage: mockStorage,
		MockOptions: mockOptions,
		MockWorker:  mockWorker,
		MockAuthz:   mockAuthz,
	}
}
func TestShortenJSON(t *testing.T) {
	testContext := NewTestContext(t)
	testCases := []struct {
		name      string
		request   *pb.ShortenJSONRequest
		expectErr bool
	}{
		{
			name: "SuccessfulShortenJSON",
			request: &pb.ShortenJSONRequest{
				Url: "http://example.com",
			},
			expectErr: false,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			// Pass the token to the context for authentication
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", testContext.FakeToken))

			resp, err := testContext.Server.ShortenJSON(ctx, tc.request)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestShortenURL(t *testing.T) {
	testContext := NewTestContext(t)

	testCases := []struct {
		name      string
		request   *pb.AddURLRequest
		expectErr bool
	}{
		{
			name: "SuccessfulShortenURL",
			request: &pb.AddURLRequest{
				Fullurl: "http://example.com",
			},
			expectErr: false,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", testContext.FakeToken))

			resp, err := testContext.Server.ShortenURL(ctx, tc.request)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestGetUserURLs(t *testing.T) {
	testContext := NewTestContext(t)

	// Mock data for the test
	mockUserID := "mockUserID"
	mockUserURLs := []models.DataURLite{
		{ShortURL: "mockShortURL1", OriginalURL: "http://example1.com"},
		{ShortURL: "mockShortURL2", OriginalURL: "http://example2.com"},
	}

	// Set up the mockStorage behavior
	testContext.MockStorage.getUserURLsFunc = func(userID string) []models.DataURLite {
		assert.Equal(t, mockUserID, userID)
		return mockUserURLs
	}

	// Perform the test
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", testContext.FakeToken))
	resp, err := testContext.Server.GetUserURLs(ctx, &pb.GetUserURLsRequest{})

	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, len(mockUserURLs), len(resp.Urls))
	// Additional assertions based on your specific response structure
}

func TestShortenBatch(t *testing.T) {
	testContext := NewTestContext(t)
	testCases := []struct {
		name      string
		request   *pb.ShortenBatchRequest
		expectErr bool
	}{
		{
			name: "SuccessfulShortenBatch",
			request: &pb.ShortenBatchRequest{
				Urls: []*pb.UrlToShorten{
					{Uuid: "1", OriginalUrl: "http://example.com"},
					{Uuid: "2", OriginalUrl: "http://example.org"},
				},
			},
			expectErr: false,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Pass the token to the context for authentication
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", testContext.FakeToken))

			// Set up a function to emulate the behavior of the ShortURLAdress method
			testContext.MockOptions.shortURLAddressFunc = func() string {
				return "mockedShortURLAddress"
			}

			// Set up a function to emulate the behavior of the  InsertBatch
			testContext.MockStorage.insertBatchFunc = func(data map[string]models.DataURL) error {
				// Code for checking and/or emulating method behavior InsertBatch
				return nil
			}

			resp, err := testContext.Server.ShortenBatch(ctx, tc.request)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	testContext := NewTestContext(t)
	testCases := []struct {
		name      string
		expectErr bool
	}{
		{
			name:      "StorageAvailable",
			expectErr: false,
		},
		{
			name:      "StorageUnavailable",
			expectErr: false,
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up a function to emulate the behavior of the  GetBaseConnection
			testContext.MockStorage.getBaseConnectionFunc = func() bool {
				// Code to emulate storage availability or unavailability
				return tc.name == "StorageAvailable"
			}

			resp, err := testContext.Server.HealthCheck(context.Background(), &pb.HealthCheckRequest{})

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				// Checks the Status field in the response
				switch tc.name {
				case "StorageAvailable":
					assert.Equal(t, pb.HealthCheckResponse_OK, resp.Status)
				case "StorageUnavailable":
					assert.Equal(t, pb.HealthCheckResponse_ERROR, resp.Status)
				}
			}
		})
	}
}

func TestGetFullURL(t *testing.T) {
	testContext := NewTestContext(t)
	testCases := []struct {
		name        string
		request     *pb.GetURLRequest
		expectErr   bool
		expectedURL string
	}{
		{
			name:        "ValidKey",
			request:     &pb.GetURLRequest{Key: "validKey"},
			expectErr:   false,
			expectedURL: "http://example.com",
		},
		{
			name:      "EmptyKey",
			request:   &pb.GetURLRequest{Key: ""},
			expectErr: true,
		},
		{
			name:      "KeyNotFound",
			request:   &pb.GetURLRequest{Key: "nonExistentKey"},
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up a function to emulate the behavior of the method GetURL
			testContext.MockStorage.getURLFunc = func(key string) (models.DataURL, error) {
				if key == "validKey" {
					return models.DataURL{OriginalURL: "http://example.com"}, nil
				}
				// Return an error if the key does not exist or other scenarios
				return models.DataURL{}, errors.New("URL not found")
			}

			// Pass the token to the context for authentication
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", testContext.FakeToken))
			resp, err := testContext.Server.GetFullURL(ctx, tc.request)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				// Check the expected URL in the response
				assert.Equal(t, tc.expectedURL, resp.OriginalUrl)
			}
		})
	}
}

// TestDeleteUserURLs tests the DeleteUserURLs method of the UsersServer.
func TestDeleteUserURLs(t *testing.T) {
	// Create a new test context
	testContext := NewTestContext(t)

	// Define test cases
	testCases := []struct {
		name      string
		request   *pb.DeleteUserURLsRequest
		expectErr bool
	}{
		{
			name:      "ValidRequest",
			request:   &pb.DeleteUserURLsRequest{Urls: []string{"url1", "url2"}},
			expectErr: false,
		},
		{
			name:      "EmptyRequest",
			request:   &pb.DeleteUserURLsRequest{Urls: []string{}},
			expectErr: false, // It is valid to delete an empty list of URLs
		},
		// Add more test cases as needed
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up mock authentication function in the test context
			testContext.MockAuthz.authCookieFunc = func(jwtToken, userID string) *http.Cookie {
				// Mocked authentication cookie
				return &http.Cookie{Name: "mockedCookie", Value: "mockedValue"}
			}

			// Set up the worker mock
			testContext.MockWorker.addFunc = func(task models.DeleteURL) {
				// Verify the correctness of the received task
				assert.Equal(t, "mockUserID", task.UserID)
				assert.ElementsMatch(t, tc.request.Urls, task.ShortURLs)
			}

			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", testContext.FakeToken))

			// Call the DeleteUserURLs method with the test request
			resp, err := testContext.Server.DeleteUserURLs(ctx, tc.request)

			// Check if the expected error occurred
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				// Check for successful execution
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

// TestRegisterUser tests the RegisterUser method of the UsersServer.
func TestRegisterUser(t *testing.T) {
	// Create a new test context
	testContext := NewTestContext(t)

	// Define test cases
	testCases := []struct {
		name        string
		request     *pb.RegisterUserRequest
		expectErr   bool
		expectedMsg string
	}{
		{
			name: "SuccessfulRegistration",
			request: &pb.RegisterUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "John Doe",
			},
			expectErr:   false,
			expectedMsg: "User test@example.com registered successfully",
		},
		{
			name: "DuplicateEmail",
			request: &pb.RegisterUserRequest{
				Email:    "existingUser@example.com",
				Password: "password456",
				Name:     "Jane Doe",
			},
			expectErr:   true,
			expectedMsg: "User with email existingUser@example.com already exists",
		},
		// Add more test cases as needed
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the mock storage to simulate existing or non-existing users
			testContext.MockStorage.getUserFunc = func(email string) (models.DataUser, error) {
				// Simulate an existing user for the DuplicateEmail test case
				if email == "existingUser@example.com" {
					return models.DataUser{}, nil
				}
				// Simulate a non-existing user for other test cases
				return models.DataUser{}, errors.New("not found")
			}

			// Set up the mock authorization to return a fixed hash for password
			testContext.MockAuthz.getHashFunc = func(email, password string) []byte {
				// Fixed hash for testing purposes
				return []byte("$2a$10$eJkK4X6vTGlIXqLa/6pSROAfWu0FJWtz9HcI8TTes/V4Xr6H3N1.2")
			}

			// Set up the mock storage's InsertUser method to return success or failure
			testContext.MockStorage.insertUserFunc = func(email string, data models.DataUser) (models.DataUser, error) {
				if email == "existingUser@example.com" {
					return models.DataUser{}, errors.New("duplicate key")
				}
				return data, nil
			}

			// Call the RegisterUser method with the test request
			resp, err := testContext.Server.RegisterUser(context.Background(), tc.request)

			// Check if the expected error occurred
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				// Check for successful execution
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				// Check if the response message matches the expected message
				assert.Equal(t, tc.expectedMsg, resp.Message)
			}
		})
	}
}

// TestLogin tests the Login method of the UsersServer.
func TestLogin(t *testing.T) {
	// Create a new test context
	testContext := NewTestContext(t)

	// Define test cases
	testCases := []struct {
		name          string
		request       *pb.LoginRequest
		expectErr     bool
		expectedToken string
	}{
		{
			name: "SuccessfulLogin",
			request: &pb.LoginRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectErr:     false,
			expectedToken: "mockedJWTToken", // Adjust based on your expected token
		},
		{
			name: "UserNotFound",
			request: &pb.LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password456",
			},
			expectErr:     true,
			expectedToken: "", // No token expected if user is not found
		},
		{
			name: "IncorrectPassword",
			request: &pb.LoginRequest{
				Email:    "existingUser@example.com",
				Password: "wrongPassword",
			},
			expectErr:     true,
			expectedToken: "", // No token expected if password is incorrect
		},
		// Add more test cases as needed
	}

	// Iterate through test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Set up the mock storage to simulate existing or non-existing users
			testContext.MockStorage.getUserFunc = func(email string) (models.DataUser, error) {
				// Simulate an existing user for SuccessfulLogin test case
				if email == "test@example.com" {
					return models.DataUser{
						UUID:  "mockUserID",
						Email: "test@example.com",
						Hash:  []byte("mockedHash"), // Mocked hash for testing
					}, nil
				}
				// Simulate a non-existing user for UserNotFound test case
				return models.DataUser{}, errors.New("not found")
			}

			// Set up the mock authorization to return a fixed hash for password
			testContext.MockAuthz.getHashFunc = func(email, password string) []byte {
				// Fixed hash for testing purposes
				return []byte("mockedHash")
			}

			// Call the Login method with the test request
			resp, err := testContext.Server.Login(context.Background(), tc.request)

			// Check if the expected error occurred
			if tc.expectErr {
				assert.Error(t, err)
				assert.Nil(t, resp)
			} else {
				// Check for successful execution
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				// Check if the response token matches the expected token
				assert.Equal(t, tc.expectedToken, resp.Token)
			}
		})
	}
}
