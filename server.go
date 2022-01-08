package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"

	"triangula-api-server/logic"
)

type Settings struct {
	points     uint
	shape      string
	mutations  uint
	variation  float64
	population uint
	cache      uint
	block      uint
	cutoff     uint
	reps       uint
	threads    uint
}

type Image struct {
	Name        string
	Hash        string
	Processed   bool
	CallbackUrl string
	Settings    Settings
}

type imageHandlers struct {
	store map[string]Image
	sync.Mutex
	jobs chan int
}

func newSettings() *Settings {
	return &Settings{
		points:     300,
		shape:      "triangles",
		mutations:  2,
		variation:  0.3,
		population: 400,
		cache:      22,
		block:      5,
		cutoff:     1,
		reps:       100,
		threads:    0,
	}
}

func (s *Settings) mergeWithPost() {

}

func (h *imageHandlers) loadStore() {
	// read images directory
	files, err := ioutil.ReadDir("images")
	if err != nil {
		panic(err.Error())
	}

	// create storage
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
		hash := md5.New()

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

	// read json and add additional data
	jsonFile, err := os.Open("storage.json")
	if err != nil {
		panic(err.Error())
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	var savedData map[string]Image
	json.Unmarshal(byteValue, &savedData)

	fmt.Print((savedData))

}

func (h *imageHandlers) saveStorage(jobs <-chan int) {

	h.Lock()
	jsonBytes, err := json.Marshal(h.store)
	if err != nil {
		panic(err.Error())
	}
	h.Unlock()

	_ = ioutil.WriteFile("storage.json", jsonBytes, 0644)
}

func (h *imageHandlers) get(w http.ResponseWriter, r *http.Request) {
	images := make([]Image, len(h.store))
	i := 0

	h.Lock()
	for _, image := range h.store {
		images[i] = image
		i++
	}
	h.Unlock()

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
	if r.Form.Has("md5") {
		if image, ok := h.store[r.Form.Get("md5")]; ok {
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

	newFile, err := os.Open("images/" + handler.Filename)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()

	// calc hash
	hash := md5.New()
	if _, err := io.Copy(hash, newFile); err != nil {
		panic(err)
	}
	fileHash := base64.RawURLEncoding.EncodeToString(hash.Sum(nil))

	// check for existence
	if image, ok := h.store[fileHash]; ok {
		if image.Processed {
			http.Redirect(w, r, "/image/"+image.Hash, http.StatusSeeOther)
			return
		}
	} else {

		// create settings
		settings := newSettings()
		if r.Form.Has("settings") {
			settings.mergeWithPost()
		}

		// add to store
		img := Image{
			Name:      handler.Filename,
			Hash:      fileHash,
			Processed: false,
			Settings:  *settings,
		}
		h.Lock()
		h.store[img.Hash] = img
		h.Unlock()

		// save store
		go h.saveStorage(h.jobs)
	}

	resp := make(map[string]string)
	resp["message"] = "File queued"
	resp["hash"] = fileHash
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		panic(err.Error())
	}
	w.Header().Add("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResp)
}

func (h *imageHandlers) startProcessing(jobs <-chan int) {

	for {
		h.Lock()
		for index, image := range h.store {
			if !image.Processed {
				h.store[index] = processFile(image)
			}
		}
		h.Unlock()
	}

}

func processFile(storeImage Image) Image {
	// calculation arguments
	image := "images/" + storeImage.Name
	json := "images/tmp.json"
	output := "images/" + storeImage.Hash
	points := 300
	shape := "triangles"
	mutations := 2
	variation := 0.3
	population := 400
	cache := 22
	block := 5
	cutoff := 1
	reps := 100
	threads := 0

	// generate json
	logic.RunAlgorithm(image, json, uint(points), shape,
		uint(mutations), float64(variation), uint(population), uint(cache),
		uint(cutoff), uint(block), uint(reps), uint(threads))

	// generate svg
	logic.RenderSVG(json, output, image, shape)

	// delete json + image
	os.Remove(json)
	os.Remove(image)

	// update image in store
	storeImage.Processed = true
	storeImage.Name = storeImage.Hash + ".svg"

	return storeImage
}

func newImageHandlers() *imageHandlers {
	return &imageHandlers{
		store: map[string]Image{},
	}
}

func main() {
	imageHandlers := newImageHandlers()
	imageHandlers.loadStore()

	go imageHandlers.saveStorage(imageHandlers.jobs)

	go imageHandlers.startProcessing(imageHandlers.jobs)

	http.HandleFunc("/", imageHandlers.get)
	http.HandleFunc("/image", imageHandlers.post)
	http.HandleFunc("/image/", imageHandlers.show)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}
