package commands

import "vc/workdir"

// VC (Version Control) represents a simplified version control system,
// similar to Git. It keeps track of a working directory and provides
// operations like add, commit, status, and checkout.
type VC struct {
	// wd stores a reference to the associated working directory.
	// All version control operations (e.g., commit, add, status)
	// will be applied to this WorkDir.
	wd *workdir.WorkDir
}

// Init initializes and returns a new VC (Version Control) instance.
// It takes a WorkDir as input and sets it as the working directory
// that this VC will manage.
func Init(w *workdir.WorkDir) *VC {
	return &VC{
		wd: w, // assign the provided WorkDir to this VC
	}
}

// GetWorkDir returns the WorkDir currently managed by this VC.
// This allows external code (like tests) to access and manipulate
// the working directory associated with the VC instance.
func (v *VC) GetWorkDir() *workdir.WorkDir {
	return v.wd
}
