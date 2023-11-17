package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShorten(t *testing.T) {

	option := NewOptions()
	option.ParseFlags()

	k, v := "testEnvKey", "testEnvValue"

	os.Setenv(k, v)
	t.Run("test func GetAsString", func(t *testing.T) {
		assert.Equal(t, GetAsString(k, ""), v, "the returned result does not match the expected one")
	})

	randomKey := "047012de-7ea9-417d-9799-6c4a5ed8df7a"
	t.Run("test func GetAsString default", func(t *testing.T) {
		assert.Equal(t, GetAsString(randomKey, "defaultValue"), "defaultValue", "the returned result does not match the expected one")
	})

	testCases := []struct {
		name     string
		testfunc func() string
		result   string
	}{
		{name: "test func RunAddr", testfunc: option.RunAddr, result: ":8080"},
		{name: "test func ShortURLAdress", testfunc: option.ShortURLAdress, result: "http://localhost:8080/"},
		{name: "test func LogLevel", testfunc: option.LogLevel, result: "info"},
		{name: "test func FileStoragePath", testfunc: option.FileStoragePath, result: "/tmp/short-url-db.json"},
		{name: "test func DataBaseDSN", testfunc: option.DataBaseDSN, result: "user=tinyurl password=example dbname=tinyurl"},
		{name: "test func JWTSigningKey", testfunc: option.JWTSigningKey, result: "test_key"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.testfunc(), tc.result, "the returned result does not match the expected one")
		})
	}
}
