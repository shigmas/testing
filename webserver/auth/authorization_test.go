package auth

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOperationsHandler(t *testing.T) {
	mux := http.NewServeMux()
	assert.NoError(t, InstallHandlers(mux), "failed to install handlers")
}

type mockWriter struct {
	header     map[string][]string
	headers    []int
	bodyBuffer bytes.Buffer
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		header:  make(map[string][]string),
		headers: make([]int, 0),
	}
}

func (w *mockWriter) Header() http.Header {
	return w.header
}

func (w *mockWriter) Write(data []byte) (int, error) {
	w.bodyBuffer.Write(data)
	return len(data), nil
}

func (w *mockWriter) WriteHeader(statusCode int) {
	w.headers = append(w.headers, statusCode)
}

func (w *mockWriter) findHeader(code int) bool {
	for _, h := range w.headers {
		if h == code {
			return true
		}
	}
	return false
}

func TestAuth(t *testing.T) {
	testCases := []struct {
		name         string
		headerVal    string
		urlHeader    string
		expectHeader int
	}{
		{"defaultFailure", "", "", http.StatusUnauthorized},
		{"defaultSuccess", defaultAuthVal, "", http.StatusAccepted},
		{"customFailure", "badauth", "expectedAuth", http.StatusUnauthorized},
		{"customSuccess", "expectedAuth", "expectedAuth", http.StatusAccepted},
	}
	for _, tCase := range testCases {
		handler := authorizationHandler{}
		t.Run(tCase.name, func(t *testing.T) {
			mockWriter := newMockWriter()
			req := http.Request{}
			req.Header = map[string][]string{authHeader: {tCase.headerVal}}
			url := url.URL{
				RawPath:  fmt.Sprintf("%s?%s=%s", Base, authHeader, tCase.urlHeader),
				RawQuery: fmt.Sprintf("%s=%s", authHeader, tCase.urlHeader),
			}
			req.URL = &url

			handler.ServeHTTP(mockWriter, &req)
			assert.True(t, mockWriter.findHeader(tCase.expectHeader), "couldn't find expected header code")
		})
	}
}
