package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/gorilla/mux"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
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
			dst, _ := os.Create("./storage/" + part.FileName())
			defer dst.Close()
			io.Copy(dst, part)
		}
	}
}

func main() {
	r := mux.NewRouter()

	fs := http.FileServer(http.Dir("./storage/"))
	markdown := http.FileServer(http.Dir("./markdown/"))
	web := http.FileServer(http.Dir("./web/"))

	fileMatcher := regexp.MustCompile(`\.[a-zA-Z]*$`)

	r.HandleFunc("/upload/", uploadHandler).Methods("POST")
	r.PathPrefix("/blobs/").Handler(http.StripPrefix("/blobs/", fs)).Methods("GET")
	r.PathPrefix("/markdowns/").Handler(http.StripPrefix("/markdowns/", markdown)).Methods("GET")
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !fileMatcher.MatchString(r.URL.Path) {
			http.ServeFile(w, r, "./web/index.html")
		} else {
			web.ServeHTTP(w, r)
		}
	})

	http.ListenAndServe(":8080", r)
}
