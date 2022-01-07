package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"triangula-api-server/logic"
)

type Image struct {
	Name      string
	Hash      string
	Processed bool
}

type imageHandlers struct {
	store map[string]Image
}

func (h *imageHandlers) loadStore() {
	// read images directory
	files, err := ioutil.ReadDir("images")
	if err != nil {
		panic(err.Error())
	}

	for _, file := range files {
		// open file
		f, err := os.Open("images/" + file.Name())
		if err != nil {
			panic(err.Error())
		}

		// skip directories, .* and .json files
		fileNameParts := strings.Split(file.Name(), ".")
		if fileNameParts[0] == "" || file.IsDir() || fileNameParts[1] == "json" {
			continue
		}

		// init hash
		hash := sha256.New()

		// get hash from filename (svg) or calculation
		fileHash := fileNameParts[0]
		fileIsProcessed := true
		if fileNameParts[1] != "svg" {
			if _, err := io.Copy(hash, f); err != nil {
				panic(err)
			}
			fileHash = base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
			fileIsProcessed = false
		}

		// construct single image response
		img := Image{
			Name:      file.Name(),
			Hash:      fileHash,
			Processed: fileIsProcessed,
		}
		h.store[img.Hash] = img
	}
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonBytes)
}

func (h *imageHandlers) show(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.String(), "/")

	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Print(parts[2])

	fmt.Print("done")
}

func (h *imageHandlers) post(w http.ResponseWriter, r *http.Request) {
	// parse form
	r.ParseMultipartForm(32 << 20)

	// check if already done from transmitted hash
	if r.Form.Has("sha256") {
		if image, ok := h.store[r.Form.Get("sha256")]; ok {
			if image.Processed {
				http.Redirect(w, r, "/image/"+image.Hash, http.StatusSeeOther)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// read file
	file, handler, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer file.Close()

	// create file
	f, err := os.OpenFile("images/"+handler.Filename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	io.Copy(f, file)

	// add to store
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		panic(err)
	}
	fileHash := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))
	img := Image{
		Name:      f.Name(),
		Hash:      fileHash,
		Processed: false,
	}
	h.store[img.Hash] = img

	fmt.Print(img)
}

func (h *imageHandlers) startProcessing() {

	i := 0
	for _, image := range h.store {
		if !image.Processed {
			processFile(&image)
		}
		i++
	}

}

func processFile(storeImage *Image) {

	// calculation arguments
	image := "images/" + storeImage.Name
	fileNameParts := strings.Split(storeImage.Name, ".")
	output := "images/" + fileNameParts[0]
	points := 300
	shape := "triangles"
	mutations := 2
	variation := 0.3
	population := 400
	cache := 22
	block := 5
	cutoff := 1
	reps := 500
	threads := 0

	// generate json
	logic.RunAlgorithm(image, output, uint(points), shape,
		uint(mutations), float64(variation), uint(population), uint(cache),
		uint(cutoff), uint(block), uint(reps), uint(threads))

	// generate svg
	input := "images/" + fileNameParts[0] + ".json"
	logic.RenderSVG(input, output, image, shape)

	// delete json + image
	os.Remove(input)
	os.Remove(image)

	storeImage.Processed = true
}

func newImageHandlers() *imageHandlers {
	return &imageHandlers{
		store: map[string]Image{},
	}
}

func main() {
	imageHandlers := newImageHandlers()
	imageHandlers.loadStore()
	imageHandlers.startProcessing()
	http.HandleFunc("/", imageHandlers.get)
	http.HandleFunc("/image", imageHandlers.post)
	http.HandleFunc("/image/", imageHandlers.show)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
