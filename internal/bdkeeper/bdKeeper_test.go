package bdkeeper

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/wurt83ow/tinyurl/internal/config"
	"github.com/wurt83ow/tinyurl/internal/logger"
	"github.com/wurt83ow/tinyurl/internal/models"
	"github.com/wurt83ow/tinyurl/internal/storage"
)

// BDKeeperSuite is a suite for testing BDKeeper.
type BDKeeperSuite struct {
	suite.Suite
	keeper *BDKeeper
}

// SetupSuite initializes the test suite by creating a connection to the test database.
func (suite *BDKeeperSuite) SetupSuite() {
	// Set up your test database connection here.
	// Use a separate test database to avoid affecting the real data.
	// Parse command line flags and environment variables for configuration options
	option := config.NewOptions()
	option.ParseFlags()
	err := option.LoadFromConfigFile("../../configs/config_test.json")
	if err != nil {
		suite.T().Fatal(err)
	}

	// Initialize logger
	nLogger, err := logger.NewLogger(option.LogLevel())
	if err != nil {

		suite.T().Fatal(err)
	}

	// Initialize storage keeper based on configuration
	suite.keeper = NewBDKeeper(option.DataBaseDSN, nLogger)
}

// SetupTest is called before each test to ensure a clean state.
func (suite *BDKeeperSuite) SetupTest() {
	// Clear the test database or perform any setup needed before each test.
	suite.clearDB()
}

// TestLoad tests the Load method of BDKeeper.
func (suite *BDKeeperSuite) TestLoad() {
	// Your test logic for Load method.
	data, err := suite.keeper.Load()
	suite.NoError(err)
	suite.NotNil(data)
}

// TestLoadUsers tests the LoadUsers method of BDKeeper.
func (suite *BDKeeperSuite) TestLoadUsers() {
	// Your test logic for LoadUsers method.
	users, err := suite.keeper.LoadUsers()
	suite.NoError(err)
	suite.NotNil(users)
}

// TestUpdateBatch tests the UpdateBatch method of BDKeeper.
func (suite *BDKeeperSuite) TestUpdateBatch() {
	// Your test logic for UpdateBatch method.
	data := []models.DeleteURL{
		{ShortURLs: []string{"short_url_1", "short_url_2"}, UserID: "user_id_1"},
		// Add more test data as needed
	}
	err := suite.keeper.UpdateBatch(data...)
	suite.NoError(err)
}

// TestSave tests the Save method of BDKeeper.
func (suite *BDKeeperSuite) TestSave() {
	// Your test logic for Save method.
	key := "test_key"
	data := models.DataURL{
		ShortURL:    "short_url",
		OriginalURL: "original_url",
		UserID:      "user_id",
		DeletedFlag: false,
	}
	result, err := suite.keeper.Save(key, data)
	suite.NoError(err)
	suite.NotNil(result)
}

// TestSaveUser tests the SaveUser method of BDKeeper.
func (suite *BDKeeperSuite) TestSaveUser() {
	// Your test logic for SaveUser method.
	key := "test_key"
	data := models.DataUser{
		Email: "test@example.com",
		Hash:  []byte("hashed_password"),
		Name:  "Test User",
	}
	result, err := suite.keeper.SaveUser(key, data)
	suite.NoError(err)
	suite.NotNil(result)
}

// TestSaveBatch tests the SaveBatch method of BDKeeper.
func (suite *BDKeeperSuite) TestSaveBatch() {
	suite.clearDB()
	// Your test logic for SaveBatch method.
	data := storage.StorageURL{
		"key_1": {UUID: "key_url_1", ShortURL: "short_url_1", OriginalURL: "original_url_1", UserID: "user_id_1", DeletedFlag: false},
		"key_2": {UUID: "key_url_2", ShortURL: "short_url_2", OriginalURL: "original_url_2", UserID: "user_id_2", DeletedFlag: false},
		// Add more test data as needed
	}

	err := suite.keeper.SaveBatch(data)
	suite.NoError(err)
}

// TestSaveConflict tests the Save method with a conflict scenario.
func (suite *BDKeeperSuite) TestSaveConflict() {
	key := "test_key"
	data := models.DataURL{
		ShortURL:    "short_url",
		OriginalURL: "original_url",
		UserID:      "user_id",
		DeletedFlag: false,
	}
	_, err := suite.keeper.Save(key, data)
	suite.NoError(err)

	// Attempt to save again with the same key, expecting a conflict
	_, err = suite.keeper.Save(key, data)
	suite.Equal(storage.ErrConflict, err)
}

// TestSaveUserConflict tests the SaveUser method with a conflict scenario.
func (suite *BDKeeperSuite) TestSaveUserConflict() {
	key := "test_key"
	data := models.DataUser{
		Email: "test@example.com",
		Hash:  []byte("hashed_password"),
		Name:  "Test User",
	}
	_, err := suite.keeper.SaveUser(key, data)
	suite.NoError(err)

	// Attempt to save again with the same key, expecting a conflict
	_, err = suite.keeper.SaveUser(key, data)
	suite.Equal(storage.ErrConflict, err)
}

// TestPing tests the Ping method of BDKeeper.
func (suite *BDKeeperSuite) TestPing() {
	// Your test logic for Ping method.
	result := suite.keeper.Ping()
	suite.True(result)
}

// TearDownSuite is called once after all tests in the suite have been run.
func (suite *BDKeeperSuite) TearDownSuite() {
	// Close any resources or connections used by the test suite.
	// For example, you may want to close the test database connection here.
	suite.keeper.Close()
}

func (suite *BDKeeperSuite) clearDB() {
	// Close any resources or connections used by the test suite.
	// For example, you may want to close the test database connection here.
	_, err := suite.keeper.conn.Exec("TRUNCATE dataurl, users CASCADE;")
	suite.NoError(err, "clear DB")
}

// Ensure BDKeeperSuite is run when the package is executed.
func TestBDKeeperSuite(t *testing.T) {
	suite.Run(t, new(BDKeeperSuite))
}
