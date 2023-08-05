module server

go 1.20

replace storage => ../storage

replace shorturl => ../shorturl

replace controllers => ../../../internal/controllers/

require (
	controllers v0.0.0-00010101000000-000000000000
	storage v0.0.0-00010101000000-000000000000
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/stretchr/testify v1.8.4 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	shorturl v0.0.0-00010101000000-000000000000 // indirect
)
