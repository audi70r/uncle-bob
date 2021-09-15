package checker

import (
	"fmt"
	"github.com/audi70r/go-archangel/utilities/clog"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

type PackageInfo struct {
	Path    string
	Files   []string
	Imports []string
	Level   int
}

// check if a package imports another package of a higher of similar level and throw a error result
func CheckLevels(packageMap map[string]PackageInfo, packageLevels [][]string) []clog.CheckResult {
	var results []clog.CheckResult

	for i := len(packageLevels) - 1; i >= 0; i-- {
		for _, pkg := range packageLevels[i] {
			for _, pkgImport := range packageMap[pkg].Imports {
				for a := i; a >= 0; a-- {
					if contains(packageLevels[a], pkgImport) {
						errMsg := fmt.Sprintf("%v is imported by a lower or similar level package %v", pkgImport, packageMap[pkg].Path)
						if !containsInCheckResults(results, errMsg) {
							results = append(results, clog.NewWarning(errMsg))
						}
					}
				}
			}
		}
	}

	return results
}

func SetLevels(packageMap map[string]PackageInfo) [][]string {
	var topLevelPackages []string

	// loop through all package imports of all packages
	for _, packageInfo := range packageMap {
		packageIsMentionedInImports := false

		// find a package that is not imported by any other packages, usually main
		for _, packageDetails := range packageMap {
			if contains(packageDetails.Imports, packageInfo.Path) {
				packageIsMentionedInImports = true
			}
		}
		if !packageIsMentionedInImports {
			topLevelPackages = append(topLevelPackages, packageInfo.Path)
		}
	}

	packagesByLevel := make([][]string, 0)
	packagesByLevel = append(packagesByLevel, topLevelPackages)

	// loop through imports of packages and group them by import level
	currentPackageLevel := 0
	for len(packagesByLevel[currentPackageLevel]) > 0 {
		packagesByLevel = append(packagesByLevel, make([]string, 0, 0))
		for _, packagePath := range packagesByLevel[currentPackageLevel] {

			if _, ok := packageMap[packagePath]; ok {
				for _, packageImport := range packageMap[packagePath].Imports {
					if !contains(packagesByLevel[currentPackageLevel+1], packageImport) {
						packagesByLevel[currentPackageLevel+1] = append(packagesByLevel[currentPackageLevel+1], packageImport)
					}
				}
			}
		}

		currentPackageLevel++
	}

	return packagesByLevel
}

func Map(workdir string) (map[string]PackageInfo, []clog.CheckResult) {
	var results []clog.CheckResult

	PackageMap := make(map[string]PackageInfo)

	filepath.Walk(workdir, func(path string, info os.FileInfo, err error) error {
		// log and skip if error is not nil
		if err != nil {
			results = append(results, clog.NewError(err.Error()))
			return nil
		}

		// skip directories, non go files and other invalid filenames
		if info.IsDir() || (len(info.Name()) > 3 && info.Name()[len(info.Name())-3:] != ".go") {
			return nil
		}

		_, fileString := filepath.Split(path)

		relPath, err := filepath.Rel(workdir, path)

		if err != nil {
			results = append(results, clog.NewError(err.Error()))
			return nil
		}

		fileImports, err := getImportsForFile(path)

		if err != nil {
			results = append(results, clog.NewError(err.Error()))
			return nil
		}

		packagePath := fmt.Sprintf("%q", ModPath+"/"+filepath.Dir(relPath))

		if _, ok := PackageMap[packagePath]; ok {
			packageMapItem := PackageMap[packagePath]
			packageMapItem.Files = append(packageMapItem.Files, fileString)
			// add missing imports
			packageImports := packageMapItem.Imports

			for _, packageImport := range fileImports {
				if strings.Index(packageImport, ModPath) > 0 {
					packageImports = AppendStringIfMissing(packageImports, packageImport)
				}
			}

			packageMapItem.Imports = packageImports
			PackageMap[packagePath] = packageMapItem
			return nil
		}

		packageInfo := PackageInfo{
			Path:  packagePath,
			Files: []string{fileString},
			Level: 0,
		}

		var packageImports []string
		for _, packageImport := range fileImports {
			if strings.Index(packageImport, ModPath) > 0 {
				packageImports = AppendStringIfMissing(packageImports, packageImport)
			}
		}

		packageInfo.Imports = packageImports
		PackageMap[packagePath] = packageInfo

		return nil
	})

	return PackageMap, results
}

func AppendStringIfMissing(slice []string, i string) []string {
	for _, ele := range slice {
		if ele == i {
			return slice
		}
	}
	return append(slice, i)
}

func getImportsForFile(path string) ([]string, error) {
	fpath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	imports, err := parser.ParseFile(token.NewFileSet(), fpath, nil, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}

	dependencies := make([]string, 0, len(imports.Imports))

	for _, v := range imports.Imports {
		dependencies = append(dependencies, v.Path.Value)
	}

	return dependencies, nil
}
