// A simple web application that allows you to shorten long URLs.
// Shortening history is stored separately for each registered user
// optional: in the postresql database or in the json file.
//
// # See also
//
// https://github.com/wurt83ow/tinyurl/
package main

import (
	"fmt"

	"github.com/wurt83ow/tinyurl/internal/app"
)

// buildVersion, buildDate, and buildCommit are global variables to store build information.
// You can use the following build command to set variable values:
// go build -ldflags="-X main.buildVersion=1.0.0 -X main.buildDate=2023-01-01 -X main.buildCommit=abc123" -o shortener
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

	if err := app.Run(); err != nil {
		panic(err)
	}
}

// getOrDefault returns the provided value or a default value if the provided value is empty.
func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}
