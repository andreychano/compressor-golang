package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMaxUploadSize_AllowsUnderLimit(t *testing.T) {
	// next handler that reads the body and responds with 200 OK
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("unexpected read error: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})

	// limit is 10 bytes
	handler := MaxUploadSize(10, next)

	body := bytes.Repeat([]byte("a"), 5)
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestMaxUploadSize_TooLargeBody(t *testing.T) {
	// next handler simulates your upload handler logic
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		if err != nil {
			// in real code you would check here
			// strings.Contains(err.Error(), "http: request body too large")
			http.Error(w, "file too large", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// limit is 10 bytes
	handler := MaxUploadSize(10, next)

	body := bytes.Repeat([]byte("a"), 20) // exceeds the limit
	req := httptest.NewRequest(http.MethodPost, "/upload", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected status %d, got %d", http.StatusRequestEntityTooLarge, w.Code)
	}
}
