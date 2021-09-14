package checker

import (
	"github.com/audi70r/go-archangel/utilities/clog"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

type PackageInfo struct {
	Path    string
	Files   []string
	Imports []string
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

		packagePath := ModPath + "/" + filepath.Dir(relPath)

		if _, ok := PackageMap[packagePath]; ok {
			packageMapItem := PackageMap[packagePath]
			packageMapItem.Files = append(packageMapItem.Files, fileString)
			// add missing imports
			packageImports := packageMapItem.Imports

			for _, packageImport := range fileImports {
				packageImports = AppendStringIfMissing(packageImports, packageImport)
			}

			packageMapItem.Imports = packageImports
			PackageMap[packagePath] = packageMapItem
			return nil
		}

		PackageMap[packagePath] = PackageInfo{
			Path:    packagePath,
			Files:   []string{fileString},
			Imports: fileImports,
		}

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
