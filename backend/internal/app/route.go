package app

import "net/http"

func SetupAppRoutes(mux *http.ServeMux, handler *AppHandler) {
	mux.HandleFunc("/", handler.Home)
}
