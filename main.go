/**
 * Achieved:
 * 1. handle create/delete events
 * 2. create/read the json file and update the map
 * 3. handle large number of file events concurrently (configurable concurrency)
 * 4. handle file rename events
 */

package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/fsnotify/fsnotify"
)

var (
	mu               sync.Mutex
	fileData         = make(map[string]int64)
	output_file_name = "fileData.json" // output file name
	path             = "./tmp"         // path to watch
	numWorkers       = 5               // number of workers to process file events
)

func saveToFile() {
	file, err := os.OpenFile(output_file_name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening fileData.json: %s", err)
		return
	}
	defer file.Close()

	os.WriteFile(output_file_name, []byte{}, 0644) // clear the file

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "\t")
	err = encoder.Encode(fileData)
	if err != nil {
		log.Printf("Error encoding fileData: %s", err)
		return
	}
}

func processFile(filePath string) {
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			mu.Lock()
			delete(fileData, filePath)
			saveToFile()
			mu.Unlock()
			return
		}
		log.Printf("Error opening file %s: %s", filePath, err)
		return
	}

	defer file.Close()

	fileStat, err := file.Stat()
	if err != nil {
		log.Printf("Error getting file %s stat: %s", filePath, err)
		return
	}

	mu.Lock()
	fileData[filePath] = fileStat.Size()
	saveToFile()
	mu.Unlock()
}

func readFromDir(path string) {
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if !file.IsDir() {
			filePath := filepath.Join(path, file.Name())
			processFile(filePath)
		}
	}
}

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fileChan := make(chan string) // channel to send file path to workers

	if err != nil {
		log.Fatal(err)
	}
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		worker := NewWorker(i, fileChan)
		wg.Add(1)
		go worker.Process()
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event:", event)
				if event.Has(fsnotify.Write) ||
					event.Has(fsnotify.Create) ||
					event.Has(fsnotify.Rename) {
					log.Println("modified file:", event.Name)
					fileChan <- event.Name
				}
				if event.Has(fsnotify.Remove) {
					log.Println("deleted file:", event.Name)
					mu.Lock()
					delete(fileData, event.Name)
					saveToFile()
					mu.Unlock()
					fileChan <- event.Name
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	readFromDir(path) // read files from the directory before starting the watcher

	// Add a path.
	err = watcher.Add(path)
	if err != nil {
		log.Fatal(err)
	}

	wg.Wait()

	runtime.Goexit()

}
