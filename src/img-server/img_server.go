// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2021 Intel Corporation

package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

const (
	serverEndpoint = ":3333"
	dirName        = "./images"
)

var images []image.Image

func main() {
	err := getAllImages(dirName)
	if err != nil {
		fmt.Println("Quitting due to errors.")
		return
	}

	http.HandleFunc("/", imgHandler)
	fmt.Printf("Listening on endpoint %s\n", serverEndpoint)
	http.ListenAndServe(serverEndpoint, nil)
}

func imgHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["image"]
	if !ok {
		http.Error(w, "No image number in request", http.StatusBadRequest)
		return
	}

	inData, err := strconv.Atoi(keys[0])
	if err != nil {
		http.Error(w, "Bad image number in request", http.StatusBadRequest)
		return
	}

	imgNum := inData % len(images) // prevents index overrun

	writeImage(w, &images[imgNum])
}

// getAllImages reads all files from a given directory and converts them
// into generic images, jpeg or otherwise, in the images[] array.
func getAllImages(dirName string) error {
	files, err1 := ioutil.ReadDir(dirName)
	if err1 != nil {
		fmt.Printf("Failed to read directory %s. Error: %s\n", dirName, err1.Error())
		return err1
	}

	images = make([]image.Image, 0, len(files))
	for _, file := range files {
		fileName := filepath.Join(dirName, file.Name())
		image, err := getImageFromFilePath(fileName)
		if err != nil {
			fmt.Printf("Failed to convert file %s to image. Error: %s",
				fileName, err.Error())
		}
		images = append(images, image)
	}

	fmt.Printf("Collected %d images\n", len(images))
	return nil
}

func getImageFromFilePath(filePath string) (image.Image, error) {

	file, err1 := os.Open(filePath)
	if err1 != nil {
		return nil, err1
	}
	defer file.Close()
	image, _, err := image.Decode(file)
	return image, err
}

// writeImage encodes an image as jpeg and writes it into ResponseWriter.
func writeImage(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		log.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		log.Println("unable to write image.")
	}
}
