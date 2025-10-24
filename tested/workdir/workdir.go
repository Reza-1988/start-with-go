package workdir

import "fmt"

// you can use this library freely: "github.com/otiai10/copy"

// WorkDir represents an in-memory working directory.
// It stores file paths and their content.
type WorkDir struct {
	files map[string]string // key: file path (e.g., "src/main.go"), value: file content
}

// InitEmptyWorkDir creates and returns an empty working directory.
func InitEmptyWorkDir() *WorkDir {
	return &WorkDir{
		files: make(map[string]string),
	}
}

func (w *WorkDir) CreateFile(path string) error {
	// Check if the file already exists in the map.
	// If it does, we return an error to prevent overwriting an existing file.
	if _, ok := w.files[path]; ok {
		return fmt.Errorf("file already exists: %s", path)
	}

	// If the file does not exist, create a new entry in the map.
	// The key is the file path, and the value (file content) starts as an empty string,
	// meaning the file exists but is currently empty — just like running "touch file.txt".
	w.files[path] = ""

	// Return nil to indicate that the file was successfully created.
	return nil
}

func (w *WorkDir) CreateDir(path string) error {

}
