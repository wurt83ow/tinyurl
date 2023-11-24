// Package compress provides a set of utilities for transparently compressing
// and decompressing HTTP data using gzip encoding. It includes compressWriter and
// compressReader types that implement the http.ResponseWriter and io.ReadCloser
// interfaces, respectively, to enable compression and decompression of transmitted data.
package compress

import (
	"compress/gzip"
	"io"
	"net/http"
)

// CompressWriter implements the http.ResponseWriter interface
// and allows it to be transparent to the server compress
// transmitted data and set correct HTTP headers
type CompressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter creates a new CompressWriter using the provided http.ResponseWriter.
func NewCompressWriter(w http.ResponseWriter) *CompressWriter {
	return &CompressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header returns the http.Header from the underlying http.ResponseWriter.
func (c *CompressWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes compressed data to the underlying gzip.Writer.
func (c *CompressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader writes the HTTP status code to the underlying http.ResponseWriter,
// and sets the "Content-Encoding" header to "gzip" if the status code indicates success.
func (c *CompressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 || statusCode == 409 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the underlying gzip.Writer and sends all data from the buffer.
func (c *CompressWriter) Close() error {
	return c.zw.Close()
}

// CompressReader implements the io.ReadCloser interface and allows
// to transparently decompress the data received from the client for the server
type CompressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader creates a new CompressReader using the provided io.ReadCloser.
func NewCompressReader(r io.ReadCloser) (*CompressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &CompressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read reads decompressed data from the underlying gzip.Reader.
func (c CompressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close closes both the underlying io.ReadCloser and gzip.Reader.
func (c *CompressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}

	return c.zr.Close()
}
