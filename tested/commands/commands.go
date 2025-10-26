package commands

import (
	"fmt"
	"strconv"
	"strings"
	"vc/workdir"
)

// Status represents the output of a "git status"-like command.
// It keeps track of which files are modified (but not staged)
// and which files are staged (ready to be committed).
type Status struct {
	ModifiedFiles []string // files changed since the last commit or after staging
	StagedFiles   []string // files that have been added with Add/AddAll
}

// commit represents a single commit in the version control history.
// Each commit has a message and a complete snapshot of the WorkDir
// at the time of the commit — similar to how Git stores the full tree.
type commit struct {
	message  string           // the commit message (e.g., "initial commit")
	snapshot *workdir.WorkDir // a deep copy of all files and directories at commit time
}

// VC (Version Control) represents a simplified version control system.
// It is responsible for tracking file changes, staging files,
// creating commits, and allowing checkouts of past commits.
type VC struct {
	// wd is the current working directory being tracked by the VC.
	// All file operations (like modifying, appending, or writing)
	// happen inside this WorkDir.
	wd *workdir.WorkDir

	// commits stores all previous commits (the commit history).
	// The last element in this slice is considered the latest commit (HEAD).
	commits []commit

	// staged stores the content of files that have been "staged" (added).
	// The key is the file path, and the value is the file content at the moment of staging.
	// This mimics Git's "index" or "staging area".
	staged map[string]string
}

// Init initializes and returns a new VC (Version Control) instance.
// It takes a WorkDir (the current working directory) as input
// and prepares a new version control system to manage it.
// This function is similar to running `git init` in a project.
func Init(w *workdir.WorkDir) *VC {
	return &VC{
		wd:     w,                       // attach the given WorkDir to this VC
		staged: make(map[string]string), // initialize an empty staging area
	}
}

// GetWorkDir returns the WorkDir currently managed by this VC instance.
// This is used by the tests (and other code) to access or modify
// the files in the working directory directly.
func (v *VC) GetWorkDir() *workdir.WorkDir {
	return v.wd
}

// Status returns a git-like status report of the repository.
// It detects which files are currently staged and which files
// have been modified (either new, changed, or edited after staging).
func (v *VC) Status() *Status {
	// Initialize the result structure with empty slices
	// (never return nil slices; tests expect [] even when empty).
	st := &Status{
		ModifiedFiles: []string{},
		StagedFiles:   []string{},
	}

	// If there are no commits yet, we have nothing to compare against.
	// So both ModifiedFiles and StagedFiles remain empty.
	if len(v.commits) == 0 {
		return st
	}

	// 1. Collect all staged files
	// Every file added via Add() or AddAll() should appear here.
	for p := range v.staged {
		st.StagedFiles = append(st.StagedFiles, p)
	}

	// 2. Retrieve the last committed snapshot
	// This represents the "clean" state of the repo at the last commit.
	lastSnap := v.commits[len(v.commits)-1].snapshot

	// Get the list of all current files in the working directory.
	curFiles := v.wd.ListFilesRoot()

	// 3. Find all files that differ from the last commit
	// These can be new files or files whose content has changed.
	changed := make(map[string]bool)
	for _, p := range curFiles {
		curContent, _ := v.wd.CatFile(p)
		prevContent, err := lastSnap.CatFile(p)

		// If the file doesn't exist in the last commit (new file)
		// OR the content is different → mark as changed.
		if err != nil || prevContent != curContent {
			changed[p] = true
		}
	}

	// 4. Determine which changed files are modified vs staged
	for p := range changed {
		curContent, _ := v.wd.CatFile(p)

		if stagedContent, isStaged := v.staged[p]; isStaged {
			// The file is already staged — check if it changed again.
			// If its content differs from the staged version,
			// it means it was modified *after* being staged.
			if curContent != stagedContent {
				st.ModifiedFiles = append(st.ModifiedFiles, p)
			}
			// If the staged version matches the current version,
			// it's not "modified" — it's cleanly staged.
		} else {
			// The file is not staged but changed ⇒ mark as modified.
			st.ModifiedFiles = append(st.ModifiedFiles, p)
		}
	}

	// Return the computed status report.
	return st
}

// AddAll stages all files currently present in the working directory,
// similar to running `git add .`. For each file, we read its current
// content and store it in the staging area (v.staged) so that the
// staged snapshot reflects the exact state at the time of staging.
func (v *VC) AddAll() {
	for _, path := range v.wd.ListFilesRoot() {
		// Read the current content of the file from the WorkDir.
		if content, err := v.wd.CatFile(path); err == nil {
			// Stage the file by recording its content at this moment.
			v.staged[path] = content
		}
	}
}

// Commit creates a new commit with the provided message.
// It takes a full snapshot of the current WorkDir (deep clone) and
// appends it to the commit history. After a successful commit,
// the staging area is cleared (just like Git).
func (v *VC) Commit(message string) {
	// Take a complete snapshot of the current working directory.
	snapshot := v.wd.Clone()

	// Append the new commit (message + snapshot) to the history.
	v.commits = append(v.commits, commit{
		message:  message,
		snapshot: snapshot,
	})

	// Clear the staging area; files were committed.
	v.staged = make(map[string]string)
}

// Add stages the specified files by saving their current content
// into the staging area (v.staged), similar to `git add file1 file2`.
// If a path does not exist in the WorkDir, it is silently ignored.
func (v *VC) Add(path ...string) {
	for _, p := range path {
		// Read current content of the file; if it exists, stage it.
		if content, err := v.wd.CatFile(p); err == nil {
			v.staged[p] = content
		}
	}
}

// Log returns a list of commit messages in reverse chronological order,
// similar to `git log`. The most recent commit appears first.
func (v *VC) Log() []string {
	log := make([]string, 0)

	// If there are no commits, return an empty list.
	if len(v.commits) == 0 {
		return log
	}

	// Iterate over commits in reverse order (latest first).
	for i := len(v.commits) - 1; i >= 0; i-- {
		log = append(log, v.commits[i].message)
	}

	return log
}

// Checkout returns a snapshot of the WorkDir from a previous commit,
// similar to the `git checkout` command. It supports two reference syntaxes:
//
//	"~N"  → go N commits back from the latest commit (HEAD~N)
//	"^^^" → number of '^' equals how many commits to go back (HEAD^^^ = HEAD~3)
//
// Example:
//
//	"~1"  → previous commit
//	"~2"  → two commits before the last
//	"^"   → previous commit (same as "~1")
//	"^^"  → two commits before the last
//
// Returns a *cloned* WorkDir so that changes to the result do not
// affect the repository’s stored history.
func (v *VC) Checkout(symbol string) (*workdir.WorkDir, error) {
	// 1. Ensure we have commits to work with.
	if len(v.commits) == 0 {
		return nil, fmt.Errorf("no commits to checkout")
	}

	steps := 0 // how many commits to go back

	// 2. Determine how many steps to move back in history.
	switch {
	case strings.HasPrefix(symbol, "~"):
		// Handle "~N" form. Extract the number after "~".
		nStr := strings.TrimPrefix(symbol, "~")

		// Convert string to integer (e.g., "~2" → 2).
		n, err := strconv.Atoi(nStr)
		if err != nil || n < 0 {
			return nil, fmt.Errorf("invalid ref: %s", symbol)
		}

		steps = n

	case strings.Trim(symbol, "^") == "" && len(symbol) > 0:
		// Handle repeated "^" form (e.g., "^", "^^", "^^^").
		// The number of carets equals how far to go back.
		steps = len(symbol)

	default:
		// Unsupported format (anything other than "~N" or "^...").
		return nil, fmt.Errorf("unsupported ref: %s", symbol)
	}

	// 3. Calculate which commit index to check out.
	// The latest commit is at index len(v.commits)-1 (HEAD).
	// So we go backwards by 'steps'.
	idx := len(v.commits) - 1 - steps

	// Check if the calculated index is valid.
	if idx < 0 || idx >= len(v.commits) {
		return nil, fmt.Errorf("ref out of range: %s", symbol)
	}

	// 4. Return a cloned snapshot of the target commit.
	// Clone ensures that any edits made to this WorkDir do not
	// modify the stored commit snapshot in history.
	return v.commits[idx].snapshot.Clone(), nil
}
