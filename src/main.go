package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log/slog"
	"main/src/api"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func AlluvialServer(r *mux.Router, prefix string, root_path string) {
	fsys := AlluvialFileSystem{http.Dir(root_path)}
	fs := FileServer(fsys)
	router := r.PathPrefix(prefix).Subrouter()

	router.PathPrefix("/").Handler(http.StripPrefix(prefix, fs)).Methods("GET")
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private")
		reader, err := r.MultipartReader()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}

			fmt.Printf("FileName=[%s], FormName=[%s]\n", part.FileName(), part.FormName())
			if part.FileName() == "" { // this is FormData``
				data, _ := ioutil.ReadAll(part)
				fmt.Printf("FormData=[%s]\n", string(data))
			} else { // This is FileData
				dst, _ := os.Create(root_path + part.FileName())
				defer dst.Close()
				io.Copy(dst, part)
			}
		}
	}).Methods("POST").MatcherFunc(func(r *http.Request, rm *mux.RouteMatch) bool {
		return is_authorization(r)
	})
}

func main() {
	var programLevel = new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug)
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: programLevel}))
	slog.SetDefault(logger)

	r := mux.NewRouter()
	api.Init(r)
	AlluvialServer(r, "/markdowns", "./markdown/")

	web := http.FileServer(http.Dir("./web/"))
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "private")
		fpath, _ := os.Getwd()
		fpath += ("/web" + r.URL.Path)

		if _, err := os.Stat(fpath); err == nil {
			web.ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, "./web/index.html")
		}
	}).Methods("GET")

	http.ListenAndServe(":20080", r)
}
