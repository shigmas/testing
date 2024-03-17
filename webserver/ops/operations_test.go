package ops

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegexp(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		expectMatch  bool
		expectOp     string
		expectParams string
	}{
		{"noArgs", "/ops/foo", true, "foo", ""},
		{"args", "/ops/foo?blah", true, "foo", "blah"},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			parts := opsRe.FindSubmatch([]byte(tCase.path))
			if !tCase.expectMatch {
				assert.Nil(t, parts, "did not expect match")
				return
			}
			assert.Equal(t, tCase.expectOp, string(parts[1]), "capture did not match")
			assert.Equal(t, tCase.expectParams, string(parts[2]), "capture did not match")
		})
	}

}

func TestOperationsHandler(t *testing.T) {
	mux := http.NewServeMux()
	assert.NoError(t, InstallHandlers(mux), "failed to install handlers")
}

func TestGetOp(t *testing.T) {
	testCases := []struct {
		name          string
		path          string
		expectedValue Op
		expectError   bool
	}{
		{"add", "/ops/add", Add, false},
	}
	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			res, _, err := GetOp(tCase.path)
			if tCase.expectError {
				assert.Error(t, err, "expected error")
				return
			}
			assert.Equal(t, tCase.expectedValue, res, "unexpected value")
		})
	}
}

func TestGetInt(t *testing.T) {
	testCases := []struct {
		name          string
		formVals      url.Values
		key           string
		expectedValue int64
		expectError   bool
	}{
		{"no val", url.Values{}, "foo", 0, true},
		{"Expected value", url.Values{"foo": []string{"3"}}, "foo", 3, false},
		{"Not int", url.Values{"foo": []string{"int", "5"}}, "foo", 3, true},
	}
	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			res, err := getInt(tCase.formVals, tCase.key)
			if tCase.expectError {
				assert.Error(t, err, "expected error")
			} else {
				assert.Equal(t, tCase.expectedValue, res, "unexpected value")
			}
		})
	}
}

type MyResponseWriter struct {
	header     http.Header
	data       []byte
	statusCode int
}

var _ (http.ResponseWriter) = (*MyResponseWriter)(nil)

func (w *MyResponseWriter) Header() http.Header {
	return w.header
}

func (w *MyResponseWriter) Write(data []byte) (int, error) {
	w.data = append(w.data, data...)
	return len(data), nil
}

func (w *MyResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

func TestList(t *testing.T) {
	handler := operationsHandler{}
	testCases := []struct {
		name         string
		url          string
		expectStatus int
		expectData   []string
	}{
		{"noVals", "/ops/list", 400, nil},
		{"oneVal", "/ops/list?arg=1", 200, []string{"1"}},
		{"oneVal", "/ops/list?arg=2&arg=5", 200, []string{"2", "5"}},
	}
	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			writer := MyResponseWriter{header: make(http.Header)}
			url := url.URL{}
			url.Path = tCase.url
			req := http.Request{}
			req.URL = &url
			req.Method = http.MethodGet

			handler.ServeHTTP(&writer, &req)
			if tCase.expectStatus != 200 {
				assert.Equal(t, tCase.expectStatus, writer.statusCode, "unexpected status code")
				return
			}
			var returned []testListObject
			fmt.Printf("data: %s\n", string(writer.data))
			err := json.Unmarshal(writer.data, &returned)
			assert.NoError(t, err, "unexpected error in unmarshaling")
			for i := 0; i < len(tCase.expectData); i++ {
				assert.Equal(t, tCase.expectData[i], returned[i].D,
					"expected data did not match")
			}
		})
	}
}

type MyReadCloser struct {
	*bytes.Reader
}

func (mrc MyReadCloser) Close() error {
	return nil
}

func TestBuildOperation(t *testing.T) {
	h := operationsHandler{}
	testGetCases := []struct {
		name        string
		formVals    url.Values
		expectError bool
	}{
		{"missing operand", url.Values{}, true},
		{"valid", url.Values{"operand1": []string{"3"}, "operand2": []string{"4"}}, false},
	}
	for _, tCase := range testGetCases {
		t.Run(tCase.name, func(t *testing.T) {
			req := http.Request{}
			req.Method = http.MethodGet
			req.Form = tCase.formVals
			_, err := h.BuildOperation("add", &req)
			if tCase.expectError {
				assert.Error(t, err, "expected error")
			}
		})
	}
	testPostCases := []struct {
		name        string
		body        string
		operand1    int
		operand2    int
		expectError bool
	}{
		{"empty body", "", 0, 0, true},
		{"not json", "<this is a test>", 0, 0, true},
		{"not json", "{\"operand1\":3,\"operand2\":5}", 3, 5, false},
	}
	for _, tCase := range testPostCases {
		t.Run(tCase.name, func(t *testing.T) {
			req := http.Request{}
			req.Method = http.MethodPost
			body := MyReadCloser{bytes.NewReader([]byte(tCase.body))}
			req.Body = body
			_, err := h.BuildOperation("add", &req)
			if tCase.expectError {
				assert.Error(t, err, "expected error")
			}
		})
	}
}
