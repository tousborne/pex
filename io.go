package main

import (
	"bufio"
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

type imageBuffer struct {
	url   string
	bytes []byte
}

// Download an image from the given url and push the image data and url to the buffer
// channel.
func downloadImage(syncro syncGroup, dispatch *sync.WaitGroup, url string) {
	defer dispatch.Done()

	// Remove the "block from the download slots to all more downloaders to run.
	defer func() {
		<-syncro.downSlots
	}()

	response, err := http.Get(url)
	if err != nil {
		syncro.err <- fmt.Errorf("Error downloading image: %s", err)
		return
	}

	buffer := imageBuffer{
		url: url,
	}

	buffer.bytes, err = ioutil.ReadAll(response.Body)
	if err != nil {
		syncro.err <- fmt.Errorf("Error downloading image: %s", err)
		return
	}

	syncro.buffer <- buffer
}

// Process the given image data and push the processed csv data to the csvs channel.
func processImage(syncro syncGroup, dispatch *sync.WaitGroup, data imageBuffer) {
	defer dispatch.Done()

	// Remove the "block" from the buffer to allow more images to be processed.
	defer func() {
		<-syncro.execSlots
	}()

	image, _, err := image.Decode(bytes.NewReader(data.bytes))
	if err != nil {
		syncro.err <- err
		return
	}

	r, g, b := parseImage(image)
	syncro.csvs <- []string{data.url, r, g, b}
}

// Read the given input file and stream the urls to the url channel.
func reader(syncro syncGroup, inPath string) {
	defer syncro.waitGroup.Done()

	inFile, err := os.Open(inPath)
	if err != nil {
		syncro.err <- fmt.Errorf("Failed to open %s: %s", inPath, err)
		return
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)

	// Defaults to tokenizing on newlines
	for scanner.Scan() {
		url := scanner.Text()
		syncro.urls <- url
	}

	close(syncro.urls)
}

// Writes any csv data recieved over the appropriate channel to the output file.
func writer(syncro syncGroup, outPath string) {
	defer syncro.waitGroup.Done()

	outFile, err := os.Create(outPath)
	if err != nil {
		syncro.err <- fmt.Errorf("Failed to open %s: %s", outPath, err)
		return
	}
	defer outFile.Close()

	csvWriter := csv.NewWriter(outFile)

	for data := range syncro.csvs {
		err = csvWriter.Write(data)
		if err != nil {
			syncro.err <- fmt.Errorf("Error writing to file: %s", err)
			return
		}
		csvWriter.Flush()
	}
}
