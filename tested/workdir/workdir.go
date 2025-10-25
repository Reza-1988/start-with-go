package workdir

import (
	"fmt"
	"strings"
)

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

// ListFilesRoot returns the list of all file paths stored in the WorkDir.
// The result includes all files (with their relative paths) in any subdirectory.
func (w *WorkDir) ListFilesRoot() []string {
	listFiles := make([]string, 0, len(w.files)) // initialize slice with enough capacity
	for k := range w.files {
		listFiles = append(listFiles, k)
	}
	return listFiles
}

// ListFilesIn returns all file paths that are under the given root directory,
// recursively (e.g., "src", returns "src/main.go", "src/workdir/file1.go", ...).
// It returns an error if the directory doesn't exist.
func (w *WorkDir) ListFilesIn(root string) ([]string, error) {
	// Ensure the directory exists; otherwise, return an error.
	if _, ok := w.dirs[root]; !ok {
		return nil, fmt.Errorf("directory does not exist: %s", root)
	}
	res := make([]string, 0)
	prefix := root + "/"

	// Pick every file whose path starts with "root/".
	// This naturally includes files in subdirectories (recursive behavior).
	for path := range w.files {
		if strings.HasPrefix(path, prefix) {
			res = append(res, path)
		}
	}
	return res, nil
}

// CatFile returns the content of a file with the given path.
// If the file does not exist in the WorkDir, it returns an error.
func (w *WorkDir) CatFile(file string) (string, error) {
	// Try to get the file content from the map.
	content, ok := w.files[file]
	if !ok {
		// Return an error if the file doesn't exist.
		return "", fmt.Errorf("file does not exist: %s", file)
	}

	// Return the file content and no error.
	return content, nil
}

// AppendToFile adds new content to the end of an existing file.
// If the file does not exist, it returns an error.
func (w *WorkDir) AppendToFile(file string, newContent string) error {
	oldContent, ok := w.files[file]
	if !ok {
		return fmt.Errorf("file does not exist: %s", file)
	}

	// Update the map with the new (concatenated) content
	w.files[file] = oldContent + newContent

	return nil
}
