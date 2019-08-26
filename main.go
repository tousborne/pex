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
	buffer    chan imageBuffer
	done      chan struct{}
	downSlots chan struct{}
	execSlots chan struct{}
	err       chan error
	csvs      chan []string
	urls      chan string
	waitGroup *sync.WaitGroup
}

// Read all urls from the url channel and dispatch their downloading routines, using the
// syncro.downSlots channel as a fake "goroutine pool".
func downDispatch(syncro syncGroup) {
	defer syncro.waitGroup.Done()

	// Dispatch waitgroup used to synchronize downloading routines.
	downGroup := new(sync.WaitGroup)
	defer close(syncro.buffer)
	defer downGroup.Wait()

	for url := range syncro.urls {
		// Use the syncro.buffer channel to limit the number of images currently
		// in memory via download / processing.
		syncro.downSlots <- struct{}{}
		downGroup.Add(1)
		go downloadImage(syncro, downGroup, url)
	}

}

// Read all available image data from buffer channel and dispatch their processing
// routines, using the syncro.execSlots channel as a fake "goroutine pool".
func execDispatch(syncro syncGroup) {
	defer syncro.waitGroup.Done()

	// Dispatch waitgroups. After all dispatch routines are done close channels.
	execGroup := new(sync.WaitGroup)
	defer close(syncro.csvs)
	defer close(syncro.done)
	defer execGroup.Wait()

	for data := range syncro.buffer {
		// Use the syncro.buffer channel to limit the number of images currently
		// in memory via download / processing.
		syncro.execSlots <- struct{}{}
		execGroup.Add(1)
		go processImage(syncro, execGroup, data)
	}

}

// Launch the goroutines and watch for errors.
func run(inPath, outPath string, exec, down uint64) error {
	syncro := syncGroup{
		buffer:    make(chan imageBuffer, exec),
		done:      make(chan struct{}),
		downSlots: make(chan struct{}, down),
		execSlots: make(chan struct{}, exec),
		err:       make(chan error),
		urls:      make(chan string, down),
		csvs:      make(chan []string, exec),
		waitGroup: new(sync.WaitGroup),
	}

	syncro.waitGroup.Add(4)
	go reader(syncro, inPath)
	go writer(syncro, outPath)
	go downDispatch(syncro)
	go execDispatch(syncro)
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
	var exec, down uint64
	var inPath, outPath string

	flag.StringVar(&inPath, "in", "", "The list of URLs of images to parse.")
	flag.StringVar(&outPath, "out", "", "The path to write the parsed csv data to.")
	flag.Uint64Var(&exec, "exec", 1, "The maximum number of processing goroutines.")
	flag.Uint64Var(&down, "down", 1, "The maximum number of downloading goroutines.")
	flag.Parse()

	if (inPath == "") || (outPath == "") {
		fmt.Fprintln(os.Stderr, "Both --in and --out flags are required")
		os.Exit(-1)
	}

	// XXX: Profiling
	//go func() {
	//	log.Println(http.ListenAndServe("localhost:6060", nil))
	//}()

	err := run(inPath, outPath, exec, down)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(-1)
	}
}
