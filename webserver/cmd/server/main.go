package main

import (
	"net/http"
	"webserver/ops"
	"webserver/auth"
)

func main() {
	mux := http.NewServeMux()
	ops.InstallHandlers(mux)
	auth.InstallHandlers(mux)
	http.ListenAndServe(":8001", mux)
}
