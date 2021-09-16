package checker

import (
	"github.com/audi70r/uncle-bob/utilities/clog"
	"golang.org/x/mod/modfile"
	"os"
)

var ModPath string

func LocateGoMod(targetPath string) {
	var err error

	ModPath, err = getModulePath(targetPath)

	if err != nil {
		clog.Error(err.Error())
		clog.Warning("Could not find go.mod file. Please make sure the target directory is correct or that go modules are initiated.")
		os.Exit(1)
	}
}

func getModulePath(targetPath string) (string, error) {
	gomod, modReadErr := os.ReadFile(targetPath + "/go.mod")

	if modReadErr != nil {
		return "", modReadErr
	}

	return modfile.ModulePath(gomod), nil
}
