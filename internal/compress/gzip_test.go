package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"testing"
)

// fakeCloser is a simple implementation of io.ReadCloser for testing purposes.
type fakeCloser struct {
	io.Reader
}

func (f *fakeCloser) Close() error {
	return nil
}

func TestCompressWriter(t *testing.T) {
	// Create a test response recorder
	recorder := httptest.NewRecorder()

	// Create a CompressWriter using the recorder
	cw := NewCompressWriter(recorder)

	// Write some data
	data := []byte("This is a test.")
	_, err := cw.Write(data)
	if err != nil {
		t.Errorf("Error writing data to CompressWriter: %v", err)
	}

	// Close the CompressWriter
	err = cw.Close()
	if err != nil {
		t.Errorf("Error closing CompressWriter: %v", err)
	}

	// Check if the response body is compressed
	gr, err := gzip.NewReader(recorder.Body)
	if err != nil {
		t.Errorf("Error creating gzip reader: %v", err)
	}
	defer gr.Close()

	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(gr)
	if err != nil {
		t.Errorf("Error reading decompressed data: %v", err)
	}

	// Check if the decompressed data matches the original data
	if !bytes.Equal(decompressed.Bytes(), data) {
		t.Error("Decompressed data does not match original data")
	}
}

func TestCompressReader(t *testing.T) {
	// Create a test response recorder with compressed data
	var compressedData bytes.Buffer
	gw := gzip.NewWriter(&compressedData)
	_, err := gw.Write([]byte("This is a test."))
	if err != nil {
		t.Errorf("Error writing data to gzip.Writer: %v", err)
	}
	err = gw.Close()
	if err != nil {
		t.Errorf("Error closing gzip.Writer: %v", err)
	}

	// Create a CompressReader using the compressed data
	cr, err := NewCompressReader(&fakeCloser{bytes.NewReader(compressedData.Bytes())})
	if err != nil {
		t.Errorf("Error creating CompressReader: %v", err)
	}
	defer cr.Close()

	// Read the decompressed data
	var decompressed bytes.Buffer
	_, err = decompressed.ReadFrom(cr)
	if err != nil {
		t.Errorf("Error reading decompressed data: %v", err)
	}

	// Check if the decompressed data matches the original data
	originalData := []byte("This is a test.")
	if !bytes.Equal(decompressed.Bytes(), originalData) {
		t.Error("Decompressed data does not match original data")
	}
}
