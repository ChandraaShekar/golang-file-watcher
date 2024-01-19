### Directory Watcher

This is a simple Golang program that watches a directory for changes and stores the information in json file.

#### Usage

1. Directory that is being watched is `./tmp`
2. Output file is `./fileData.json`
3. Run `go run main.go worker.go` to start the program

#### What it does?

1. Watches the directory for changes
2. Stores the file name/path, size as `{ "filePath": "size", ... }`
3. Checks if the file is created, modified or deleted
