package ops

import (
	"bytes"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			res, err := GetOp(tCase.path)
			if tCase.expectError {
				assert.Error(t, err, "expected error")
			} else {
				assert.Equal(t, tCase.expectedValue, res, "unexpected value")
			}
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

type MyReadCloser struct {
	*bytes.Reader
}

func (mrc MyReadCloser) Close() error {
	return nil
}

func TestBuildOperation(t *testing.T) {
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
			_, err := BuildOperation("add", &req)
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
			_, err := BuildOperation("add", &req)
			if tCase.expectError {
				assert.Error(t, err, "expected error")
			}
		})
	}
}
