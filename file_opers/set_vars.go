package fileopers

import "log"

// Variables used by this package.
var (
	workingDir string
	verbose    bool
)

// SetWorkingDir sets the working directory.
func SetWorkingDir(wd string) {
	workingDir = wd
	if verbose {
		log.Println("workingDir:", workingDir)
	}
}

// SetVerbose sets verbose to v.
func SetVerbose(v bool) {
	verbose = v
}
