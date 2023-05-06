package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func AlluvialServer(r *mux.Router, prefix string, root_path string) {
	fs := http.FileServer(http.Dir(root_path))
	router := r.PathPrefix(prefix).Subrouter()

	router.PathPrefix("/").Handler(http.StripPrefix(prefix, fs)).Methods("GET")
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reader, err := r.MultipartReader()
		w.Header().Add("Access-Control-Allow-Origin", "*")
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
	}).Methods("POST")
}

func main() {
	r := mux.NewRouter()
	AlluvialServer(r, "/blobs", "./storage/")
	AlluvialServer(r, "/markdowns", "./markdown/")

	web := http.FileServer(http.Dir("./web/"))

	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		fpath, _ := os.Getwd()
		fpath += ("/web/" + r.URL.Path)
		if _, err := os.Stat(fpath); err == nil {
			web.ServeHTTP(w, r)
		} else {
			log.Print(err)
			http.ServeFile(w, r, "./web/index.html")
		}

	})

	http.ListenAndServe(":8080", r)
}
