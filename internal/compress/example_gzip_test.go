// Package compress provides a set of utilities for transparently compressing
// and decompressing HTTP data using gzip encoding. It includes compressWriter and
// compressReader types that implement the http.ResponseWriter and io.ReadCloser
// interfaces, respectively, to enable compression and decompression of transmitted data.
package compress_test

import (
	"bytes"
	"fmt"
	"io"
	"net/http/httptest"

	"github.com/wurt83ow/tinyurl/internal/compress"
)

// ExampleCompressWriter demonstrates how to use the compressWriter to compress HTTP data.
func ExampleCompressWriter_Write() {
	// Create a new HTTP request with a recorder
	w := httptest.NewRecorder()

	// Create a compressWriter
	cw := compress.NewCompressWriter(w)

	// Write compressed data
	data := []byte("This is an example of compressed data.")
	_, _ = cw.Write(data)

	// Close the compressWriter to flush the data
	cw.Close()

	// Print the result
	fmt.Println(w.Header().Get("Content-Encoding"))
	fmt.Println(w.Body.String())

}

// ExampleCompressReader demonstrates how to use the compressReader to decompress HTTP data.
func ExampleCompressReader_Read() {
	// Create a new HTTP request with a recorder and a compressWriter
	recorder := httptest.NewRecorder()
	cw := compress.NewCompressWriter(recorder)

	// Write compressed data
	data := []byte("This is an example of compressed data.")
	_, _ = cw.Write(data)
	cw.Close()

	readCloser := io.NopCloser(bytes.NewReader(data))

	// Create a compressReader
	cr, err := compress.NewCompressReader(readCloser)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Read decompressed data
	result, _ := io.ReadAll(cr)

	// Print the result
	fmt.Println(string(result))

}
