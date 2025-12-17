package api

import (
	"net/http"

	"github.com/gorilla/mux"
)

const (
	_prefix = "/blobs/"
)

func Init(r *mux.Router) {
	r.PathPrefix(_prefix).Handler(
		http.StripPrefix(_prefix, http.FileServer(
			dotFileHidingFileSystem{http.Dir("./storage/")})))
}
