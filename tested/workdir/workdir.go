package workdir

import "fmt"

// you can use this library freely: "github.com/otiai10/copy"

// WorkDir represents an in-memory working directory.
// It stores file paths and their content.
type WorkDir struct {
	files map[string]string // key: file path (e.g., "src/main.go"), value: file content
	dirs  map[string]bool   // key: directory path (e.g., "src" or "src/workdir"), value: true means this directory exists
}

// InitEmptyWorkDir creates and returns an empty working directory.
func InitEmptyWorkDir() *WorkDir {
	return &WorkDir{
		files: make(map[string]string),
		dirs:  make(map[string]bool),
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

// CreateDir creates a new directory at the given path.
// It keeps track of directories in the `dirs` map to ensure they exist in memory.
// If the directory already exists, or if a file with the same name exists, it returns an error.
func (w *WorkDir) CreateDir(path string) error {
	// Check if the directory already exists in the map.
	// If it does, we return an error to prevent duplicate directories.
	if _, ok := w.dirs[path]; ok {
		return fmt.Errorf("directory already exists: %s", path)
	}

	// Check if a file with the same name already exists.
	// A path cannot represent both a file and a directory at the same time.
	if _, ok := w.files[path]; ok {
		return fmt.Errorf("a file with the same name already exists: %s", path)
	}

	// If the path is new, mark this directory as existing by setting it to true.
	w.dirs[path] = true

	// Return nil to indicate the directory was successfully created.
	return nil
}

// WriteToFile replaces the content of an existing file with new text.
// If the file does not exist, it returns an error.
func (w *WorkDir) WriteToFile(path string, content string) error {
	// Check if the file exists in the map.
	// If it doesn't, return an error to indicate that the file must be created first.
	if _, ok := w.files[path]; !ok {
		return fmt.Errorf("file does not exist: %s", path)
	}

	// Overwrite the file content with the new data.
	w.files[path] = content

	// Return nil to indicate the operation was successful.
	return nil
}

// Clone creates and returns a deep copy of the current WorkDir.
// This ensures that the cloned WorkDir is completely independent
// of the original — changes in one will not affect the other.
func (w *WorkDir) Clone() *WorkDir {
	// Initialize a new empty WorkDir to store the copied data.
	cloneWD := InitEmptyWorkDir()

	// Copy all file entries (path → content) into the new WorkDir.
	// Each key-value pair is duplicated so that maps are not shared.
	for k, v := range w.files {
		cloneWD.files[k] = v
	}

	// Copy all directory entries (path → exists) into the new WorkDir.
	for k, v := range w.dirs {
		cloneWD.dirs[k] = v
	}

	// Return the fully cloned WorkDir instance.
	return cloneWD
}
