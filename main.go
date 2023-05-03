package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
			dst, _ := os.Create("./upload/" + part.FileName())
			defer dst.Close()
			io.Copy(dst, part)
		}
	}
}

func main() {
	http.HandleFunc("/upload/", uploadHandler)

	fs := http.FileServer(http.Dir("./storage/"))
	http.Handle("/blobs/", http.StripPrefix("/blobs/", fs))

	web := http.FileServer(http.Dir("./web/"))
	http.Handle("/", http.StripPrefix("/", web))

	http.ListenAndServe(":8080", nil)
}
