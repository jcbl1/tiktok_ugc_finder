package fileopers

import "log"

var (
	workingDir string
	verbose    bool
)

func SetWorkingDir(wd string) {
	workingDir = wd
	if verbose {
		log.Println("workingDir:", workingDir)
	}
}

func SetVerbose(v bool) {
	verbose = v
}
