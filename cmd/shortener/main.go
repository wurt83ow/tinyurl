// A simple web application that allows you to shorten long URLs.
// Shortening history is stored separately for each registered user
// optional: in the postresql database or in the json file.
//
// # See also
//
// https://github.com/wurt83ow/tinyurl/
package main

//go:generate go run . "buildVersion" "1.2.3" "buildDate" "2023-01-01" "buildCommit" "abc123"

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wurt83ow/tinyurl/internal/app"
)

// buildVersion, buildDate, and buildCommit are global variables to store build information.
// You can use the following build command to set variable values:
// go build -ldflags="-X main.buildVersion=1.1.1 -X main.buildDate=2023-11-11 -X main.buildCommit=abc777" -o shortener
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// Print build information using formatted strings and default values if the variables are empty.
	fmt.Printf("Build version: %s\n", getOrDefault(buildVersion, "N/A"))
	fmt.Printf("Build date: %s\n", getOrDefault(buildDate, "N/A"))
	fmt.Printf("Build commit: %s\n", getOrDefault(buildCommit, "N/A"))

	if err := app.Run(); err != nil && err != http.ErrServerClosed {
		panic(err)
	}
	log.Printf("server stopped")
}

// getOrDefault returns the provided value or a default value if the provided value is empty.
func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
