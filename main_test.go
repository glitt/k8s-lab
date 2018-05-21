package main

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"sync/atomic"
	"testing"
)

func TestHealthCheckHandler(t *testing.T) {
	// Create a request to pass to our handler. No query parameters for now, so nil.
	req, err := http.NewRequest("GET", "/healthz", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Assign handler.
	handler := healthz()

	// Handler satisfies http.Handler.
	handler.ServeHTTP(rr, req)

	// Check the status code.
	if status := rr.Code; status != http.StatusServiceUnavailable {
		t.Errorf("Handler returned wrong status code: got %v want %v.",
			status, http.StatusServiceUnavailable)
	}

	// Check the response body.
	expected := ""
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v.",
			rr.Body.String(), expected)
	}
}

func TestIndexHandler(t *testing.T) {
	// Create a request to pass to our handler. No query parameters for now, so nil.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Assign handler.
	handler := index()

	// Handler satisfies http.Handler.
	handler.ServeHTTP(rr, req)

	// Check the status code.
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Handler returned wrong status code: got %v want %v.",
			status, http.StatusInternalServerError)
	}

	// Check response body.
	expected := "Not ready yet, World!"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v.",
			rr.Body.String(), expected)
	}

	// Set healthy to 1.
	atomic.StoreInt32(&healthy, 1)

	// Recreate recorder.
	rr = httptest.NewRecorder()

	handler = index()
	handler.ServeHTTP(rr, req)

	expected = "Hello, World!"
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v.",
			rr.Body.String(), expected)
	}
}

func TestRequestIdMiddleware(t *testing.T) {
	// Create a request to pass to our handler. No query parameters for now, so nil.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Mock request id.
	mockRequestID := func() string {
		return "111111"
	}

	// Assign handler.
	handler := tracing(mockRequestID)(index())

	// Handler satisfies http.Handler.
	handler.ServeHTTP(rr, req)

	// Get headers.
	headers := rr.Header()

	// Check the header "X-Request-Id".
	expected := mockRequestID()
	if headers.Get("X-Request-Id") != expected {
		t.Errorf("Middleware returned unexpected body: got %v want %v.",
			headers.Get("X-Request-Id"), expected)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	// Create a request to pass to our handler. No query parameters for now, so nil.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Mock request id.
	mockRequestID := func() string {
		return "111111"
	}

	logger := log.New(os.Stderr, "http: ", log.LstdFlags)

	// Set logger output to buffer.
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	defer func() {
		log.SetOutput(os.Stderr)
	}()

	// Assign handler.
	handler := tracing(mockRequestID)(logging(logger)(index()))

	// Handler satisfies http.Handler.
	handler.ServeHTTP(rr, req)

	r, err := regexp.Compile(`\s(\d+)\sGET`)

	if err != nil {
		t.Errorf("Regular expression not compiled %v.", err)
		return
	}

	// Try to match.
	resultSlice := r.FindStringSubmatch(buf.String())

	// If first group is not present bail.
	if len(resultSlice) != 2 {
		t.Errorf("Regular expression not matched %v.", err)
		return
	}

	// Check the header "X-Request-Id".
	expected := mockRequestID()

	// Compare expected to second element of slice.
	if expected != resultSlice[1] {
		t.Errorf("Middleware returned unexpected body: got %v want %v.",
			buf.String(), expected)
	}
}
