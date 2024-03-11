// Package ops is the operations url
package ops

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Base is the url base for Operations
const Base = "ops"

// Op is the type for the operation
type Op string

// These are the specific operations
const (
	Add      Op = "add"
	Subtract    = "subtract"
	Echo        = "echo"
	Error       = "error"
)

// These are the consts for the arguments in the URL or body
const (
	Opand1 = "operand1"
	Opand2 = "operand2"
	Result = "result"
)

// Operation is the container for our operaiton
type Operation struct {
	Operand1 int64 `json:"operand1"`
	Operand2 int64 `json:"operand2"`
	Oper     Op    `json:"operation"`
	Result   int64 `json:"result"`
}
type operationsHandler struct {
}

// GetOp will get the operation from the URL
func GetOp(urlPath string) (Op, error) {
	// why doesn't filepath.Splist work?
	parts := strings.Split(urlPath, "/")
	if len(parts) < 2 || parts[1] != Base {
		return "", fmt.Errorf("url path does not have leading %s", Base)
	}
	if len(parts) == 3 {
		return Op(parts[2]), nil
	}
	return "", nil
}

// InstallHandlers installs the ops hander on the mux
func InstallHandlers(mux *http.ServeMux) error {
	opsHandler := operationsHandler{}
	mux.Handle("/"+Base+"/", &opsHandler)

	return nil
}

func getInt(formVals url.Values, key string) (int64, error) {
	stringVals, ok := formVals[key]
	if !ok {
		return 0, fmt.Errorf("no operand")
	}
	if len(stringVals) != 1 {
		return 0, fmt.Errorf("invalid value for %s", key)
	}
	return strconv.ParseInt(stringVals[0], 10, 64)
}

// BuildOperation takes the op and the request and fills out the Operation
func (h *operationsHandler) BuildOperation(op Op, req *http.Request) (Operation, error) {
	oper := Operation{
		Oper: op,
	}
	val := int64(0)
	var err error
	switch req.Method {
	case http.MethodGet:
		val, err = getInt(req.Form, Opand1)
		if err != nil {
			return oper, err
		}
		oper.Operand1 = val
		val, err = getInt(req.Form, Opand2)
		if err != nil {
			return oper, err
		}
		oper.Operand2 = val
	case http.MethodPost:
		vals := struct {
			Operand1 int64 `json:"operand1"`
			Operand2 int64 `json:"operand2"`
		}{}
		decoder := json.NewDecoder(req.Body)
		if err = decoder.Decode(&vals); err != nil {
			return oper, fmt.Errorf("failed to decode body: %w", err)
		}
		oper.Operand1 = vals.Operand1
		oper.Operand2 = vals.Operand2
	default:
		return oper, nil
	}

	return oper, nil
}

// SendOperationResult Sends the result of the operation.
func (h *operationsHandler) SendOperationResult(op Operation, w http.ResponseWriter) {
	switch op.Oper {
	case Add:
		op.Result = op.Operand1 + op.Operand2
	case Subtract:
		op.Result = op.Operand1 - op.Operand2
	case Echo:
		op.Result = 0
	case Error:
		// no op. we shouldn't be here
	}

	data, err := json.Marshal(op)
	if err != nil {
		http.Error(w, "Failure in marshaling data: %s"+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("content-type", "application/json")
	w.Write(data)
}

func (h *operationsHandler) isJsonContent(req *http.Request) bool {
	ct := req.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		return mediaType == "application/json"
	}
	return false
}

func (h *operationsHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	// our sub handler.
	if !h.isJsonContent(req) {
		// ParseForm will ReadAll on the body, which is before we can unmarshal
		// JSON
		if err := req.ParseForm(); err != nil {
			http.Error(w, "Unable to parse form: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	op, err := GetOp(req.URL.Path)
	if err != nil {
		http.Error(w, "Invalid operation in "+req.URL.Path, http.StatusBadRequest)
		return
	}
	switch op {
	case Error:
		http.Error(w, "User requested error", http.StatusInternalServerError)
		return
	case Add, Subtract, Echo:
		operation, err := h.BuildOperation(op, req)
		if err != nil {
			http.Error(w, "Failure in arguments: %s"+err.Error(), http.StatusBadRequest)
		}
		h.SendOperationResult(operation, w)
	default:
		w.Write([]byte("[\"add\",\"subtract\",\"echo\",\"error\"]"))
	}
}
