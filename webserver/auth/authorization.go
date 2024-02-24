// Package auth is the authorization tester
package auth

import (
	"net/http"
)

// Base is the url base for Authorization
const Base = "auth"

const (
	authHeader     = "Authorization"
	defaultAuthVal = "default"
)

type authorizationHandler struct {
}

// InstallHandlers installs the ops hander on the mux
func InstallHandlers(mux *http.ServeMux) error {
	opsHandler := authorizationHandler{}
	mux.Handle("/"+Base+"/", &opsHandler)

	return nil
}

func (h *authorizationHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	authHeaders, ok := req.Header[authHeader]
	if !ok {
		http.Error(w, "No Auth Header", http.StatusUnauthorized)
		return
	}
	authVal := defaultAuthVal
	stringVals, ok := req.Form[authHeader]
	if ok && len(stringVals[0]) > 0 {
		authVal = stringVals[0]
	}

	if authHeaders[0] != authVal {
		http.Error(w, "No Auth Header", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("<html>Successful auth</html>"))
}
