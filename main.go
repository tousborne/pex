package main

import (
	"flag"
	"fmt"
	"os"
	"sync"
)

// XXX: profiling
//import (
//	"log"
//	"net/http"
//	_ "net/http/pprof"
//)

// Synchronization data to be passed around routines.
type syncGroup struct {
	buffer    chan struct{}
	done      chan struct{}
	err       chan error
	csvs      chan []string
	urls      chan string
	waitGroup *sync.WaitGroup
}

// Read all urls from the url channel and dispatch their processing to goroutines,
// using the syncro.buffer as a "goroutine pool".
func dispatch(syncro syncGroup) {
	defer syncro.waitGroup.Done()

	dispatchGroup := new(sync.WaitGroup)

	for url := range syncro.urls {
		// Use the syncro.buffer channel to limit the number of images currently
		// in memory via download / processing.
		syncro.buffer <- struct{}{}
		dispatchGroup.Add(1)
		go processImage(syncro, dispatchGroup, url)
	}

	dispatchGroup.Wait()
	close(syncro.csvs)
	close(syncro.done)
}

// Launch the goroutines and watch for errors.
func run(inPath, outPath string, size uint64) error {
	syncro := syncGroup{
		buffer:    make(chan struct{}, size),
		done:      make(chan struct{}),
		err:       make(chan error),
		urls:      make(chan string, size),
		csvs:      make(chan []string, size),
		waitGroup: new(sync.WaitGroup),
	}

	syncro.waitGroup.Add(3)
	go reader(syncro, inPath)
	go writer(syncro, outPath)
	go dispatch(syncro)
	defer syncro.waitGroup.Wait()

	for {
		select {
		case err := <-syncro.err:
			return err

		// syncro.csvs is the last channel to close
		case _, open := <-syncro.done:
			if !open {
				return nil
			}
		}
	}
}

// Parse the arguments and call run().
func main() {
	var size uint64
	var inPath, outPath string

	flag.StringVar(&inPath, "in", "", "The list of URLs of images to parse.")
	flag.StringVar(&outPath, "out", "", "The path to write the parsed csv data to.")
	flag.Uint64Var(&size, "size", 1, "The maximum number of images to keep buffered.")
	flag.Parse()

	if (inPath == "") || (outPath == "") {
		fmt.Fprintf(os.Stderr, "Both --in and --out flags are required")
		os.Exit(-1)
	}

	// XXX: Profiling
	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()

	err := run(inPath, outPath, size)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(-1)
	}
}
