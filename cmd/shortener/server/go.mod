module server

go 1.20

replace handler => ../handler

replace storage => ../storage

replace shorturl => ../shorturl

require (
	handler v0.0.0-00010101000000-000000000000
	storage v0.0.0-00010101000000-000000000000
)

require shorturl v0.0.0-00010101000000-000000000000 // indirect
