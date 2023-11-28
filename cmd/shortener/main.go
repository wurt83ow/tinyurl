// A simple web application that allows you to shorten long URLs.
// Shortening history is stored separately for each registered user
// optional: in the postresql database or in the json file.
//
// # See also
//
// https://github.com/wurt83ow/tinyurl/
package main

import (
	"github.com/wurt83ow/tinyurl/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		panic(err)
	}
}
