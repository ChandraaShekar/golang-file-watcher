package main

type Worker struct {
	ID    int
	Files chan string
}

func NewWorker(id int, files chan string) *Worker {
	return &Worker{
		ID:    id,
		Files: files,
	}
}

func (w *Worker) Process() {
	for filePath := range w.Files {
		processFile(filePath)
	}
}
