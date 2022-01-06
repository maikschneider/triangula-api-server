package main

import (
	"encoding/json"
	"net/http"
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
	images := make([]Image, len(h.store))

	i := 0
	for _, image := range h.store {
		images[i] = image
		i++
	}

	jsonBytes, err := json.Marshal(images)
	if err != nil {
		// TODO
	}
	w.Write(jsonBytes)
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
