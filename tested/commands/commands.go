package commands

import "vc/workdir"

type VC struct {
	wd *workdir.WorkDir
}

func Init(w *workdir.WorkDir) *VC {
	return &VC{
		wd: w,
	}
}

func (v *VC) GetWorkDir() *workdir.WorkDir {
	return v.wd
}
