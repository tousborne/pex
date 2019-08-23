package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"sync"
)

// Download an image image from the given url and process it, then send it to the csv
// channel.
func processImage(syncro syncGroup, dispatch *sync.WaitGroup, url string) {
	defer dispatch.Done()

	// Remove the "block" from the buffer to allow more images to be processed.
	defer func() {
		<-syncro.buffer
	}()

	response, err := http.Get(url)
	if err != nil {
		syncro.err <- fmt.Errorf("Error downloading image: %s", err)
		return
	}

	image, _, err := image.Decode(response.Body)
	if err != nil {
		syncro.err <- err
		return
	}

	r, g, b := parseImage(image)
	syncro.csvs <- []string{url, r, g, b}
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
