package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type Image struct {
	Name      string
	Hash      string
	Processed bool
}

type imageHandlers struct {
	store map[string]Image
}

func (h *imageHandlers) get(w http.ResponseWriter, r *http.Request) {
	// read images directory
	files, err := ioutil.ReadDir("images")
	if err != nil {
		panic(err.Error())
	}

	images := make([]Image, len(files))
	i := 0

	for _, file := range files {
		// open file
		f, err := os.Open("images/" + file.Name())
		if err != nil {
			panic(err.Error())
		}

		// calc hash
		hash := sha1.New()
		if _, err := io.Copy(hash, f); err != nil {
			panic(err)
		}

		// construct single image response
		images[i] = Image{
			Name:      file.Name(),
			Hash:      base64.URLEncoding.EncodeToString(hash.Sum(nil)),
			Processed: false,
		}
		i++

	}

	jsonBytes, err := json.Marshal(images)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *imageHandlers) post(w http.ResponseWriter, r *http.Request) {

}

func newImageHandlers() *imageHandlers {
	return &imageHandlers{
		store: map[string]Image{
			"id1": {
				Name:      "hello.jpg",
				Hash:      "fwfwefwefw11212",
				Processed: false,
			},
		},
	}
}

func main() {
	imageHandlers := newImageHandlers()
	http.HandleFunc("/images", imageHandlers.get)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
