package checker

import (
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/audi70r/uncle-bob/utilities/clog"
)

// PackageInfo represents information about a Go package
type PackageInfo struct {
	Path    string   // Package import path
	Files   []string // Go files in the package
	Imports []string // Packages imported by this package
	Level   int      // Dependency level (calculated later)
}

// HasViolations returns true if any violations were found
func HasViolations(results []clog.CheckResult) bool {
	for _, result := range results {
		if result.Type == clog.ResultWarning {
			return true
		}
	}
	return false
}

// CheckLevels verifies package imports follow Clean Architecture rules
// Returns a list of violations found with suggestions for fixes
func CheckLevels(packageMap map[string]PackageInfo, packageLevels [][]string, strict bool) []clog.CheckResult {
	var results []clog.CheckResult

	for i := len(packageLevels) - 1; i >= 0; i-- {
		for _, pkg := range packageLevels[i] {
			// Skip entry point packages (like main) - they're allowed to import from any layer
			if isEntryPointPackage(pkg) {
				continue
			}

			for _, pkgImport := range packageMap[pkg].Imports {
				for a := i; a >= 0; a-- {
					var violation bool
					var message string
					var suggestion string

					if strict {
						// In strict mode, only allow importing packages from the next level down
						if contains(packageLevels[a], pkgImport) && i-1 != a {
							violation = true
							message = "Only one level inward importing is allowed"

							if i <= a {
								// Importing from same or higher level
								suggestion = fmt.Sprintf("Consider moving '%s' to a deeper level than '%s'",
									strings.Trim(pkgImport, `"`), strings.Trim(packageMap[pkg].Path, `"`))
							} else {
								// Importing from too deep
								suggestion = fmt.Sprintf("Consider reorganizing layers to ensure '%s' is exactly one level below '%s'",
									strings.Trim(pkgImport, `"`), strings.Trim(packageMap[pkg].Path, `"`))
							}
						}
					} else {
						// In normal mode, disallow same-level or higher-level imports
						if contains(packageLevels[a], pkgImport) && i <= a {
							violation = true
							message = "Importing a package of the same level is not allowed"

							// If it's a utility package being imported at the same level
							if isUtilityPackage(pkgImport) {
								suggestion = fmt.Sprintf("'%s' appears to be a utility package. Consider moving it to a deeper level",
									strings.Trim(pkgImport, `"`))
							} else {
								suggestion = fmt.Sprintf("Consider moving '%s' to a deeper level than '%s'",
									strings.Trim(pkgImport, `"`), strings.Trim(packageMap[pkg].Path, `"`))
							}
						}
					}

					if violation {
						fromPkg := strings.Trim(packageMap[pkg].Path, `"`)
						toPkg := strings.Trim(pkgImport, `"`)

						errMsg := fmt.Sprintf("%s\nLv%d: %s <-- Lv%d: %s\nSuggestion: %s",
							message, i, fromPkg, a, toPkg, suggestion)

						if !containsInCheckResults(results, errMsg) {
							results = append(results, clog.NewWarning(errMsg))
							break
						}
					}
				}
			}
		}
	}

	// If there are violations, display them with a header
	if len(results) > 0 {
		// Display a violations header
		headerMsg := "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
		clog.PrintColorMessage(clog.NewInfo(headerMsg))

		titleMsg := "┃  🚨 DEPENDENCY VIOLATIONS                             ┃"
		clog.PrintColorMessage(clog.NewInfo(titleMsg))

		footerMsg := "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
		clog.PrintColorMessage(clog.NewInfo(footerMsg))

		// Print each violation with an enhanced visual format
		for i, v := range results {
			// Add a separator between violations
			if i > 0 {
				sepMsg := "┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄"
				clog.PrintColorMessage(clog.NewInfo(sepMsg))
			}

			clog.PrintColorMessage(v)
		}
	}

	return results
}

// CleanArchitectureLayers contains descriptions of layers in Clean Architecture
var CleanArchitectureLayers = []string{
	"Frameworks & Drivers (outermost layer)",
	"Interface Adapters",
	"Application Business Rules",
	"Enterprise Business Rules (innermost layer)",
}

// LevelsInfo displays information about package levels with Clean Architecture context
func LevelsInfo(packageLevels [][]string) {
	// Show Clean Architecture layer info with a section header
	if len(packageLevels) > 0 {
		headerMsg := "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
		clog.PrintColorMessage(clog.NewInfo(headerMsg))

		titleMsg := "┃  🏛️  CLEAN ARCHITECTURE REFERENCE                     ┃"
		clog.PrintColorMessage(clog.NewInfo(titleMsg))

		footerMsg := "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
		clog.PrintColorMessage(clog.NewInfo(footerMsg))

		// Calculate the maximum level we have, to map to Clean Architecture layers
		maxLevels := len(packageLevels)
		caLayers := CleanArchitectureLayers

		// If we have fewer packages than layers, truncate the layers to match
		if maxLevels < len(caLayers) {
			caLayers = caLayers[:maxLevels]
		}

		// Print Clean Architecture layer guidance
		for i, layer := range caLayers {
			// Use different symbols for each layer to visually distinguish them
			var layerSymbol string
			switch i {
			case 0:
				layerSymbol = "🌐" // UI/Web/External interfaces
			case 1:
				layerSymbol = "🔌" // Adapters/Controllers
			case 2:
				layerSymbol = "⚙️" // Use Cases/Application logic
			case 3:
				layerSymbol = "📦" // Entities/Domain models
			default:
				layerSymbol = "•"
			}

			levelMsg := fmt.Sprintf("%s Level %d ~ %s", layerSymbol, i, layer)
			indentedResult := clog.NewInfo(levelMsg).WithIndent(1)
			clog.PrintColorMessage(indentedResult)
		}

		clog.Info("")
	}

	// Display actual package levels with a nice header
	headerMsg := "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
	clog.PrintColorMessage(clog.NewInfo(headerMsg))

	titleMsg := "┃  📊 PACKAGE DEPENDENCY LEVELS                        ┃"
	clog.PrintColorMessage(clog.NewInfo(titleMsg))

	footerMsg := "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
	clog.PrintColorMessage(clog.NewInfo(footerMsg))

	for lvl, packageLevel := range packageLevels {
		if len(packageLevel) == 0 {
			continue
		}

		// Determine if this level aligns with a Clean Architecture layer
		layerInfo := ""
		if lvl < len(CleanArchitectureLayers) {
			layerInfo = fmt.Sprintf(" ~ %s", CleanArchitectureLayers[lvl])
		}

		// Use a visual separator between levels
		if lvl > 0 {
			separatorMsg := "┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄"
			sepResult := clog.NewInfo(separatorMsg)
			clog.PrintColorMessage(sepResult)
		}

		// Add emoji for the level to make it visually distinct
		var levelEmoji string
		switch lvl {
		case 0:
			levelEmoji = "🔝" // Top level
		case 1:
			levelEmoji = "🔼" // High level
		case 2:
			levelEmoji = "⏺️" // Mid level
		case 3:
			levelEmoji = "🔽" // Low level
		default:
			levelEmoji = "⏬" // Lowest levels
		}

		msg := fmt.Sprintf("%s Level %d%s packages:", levelEmoji, lvl, layerInfo)
		result := clog.NewInfo(msg)
		clog.PrintColorMessage(result)

		// Print each package with indentation for better readability
		for _, packageImport := range packageLevel {
			// Clean up the package string for display
			cleanPkg := strings.Trim(packageImport, `"`)

			// Determine package type label and icon
			packageLabel := ""
			packageIcon := "└─"

			if isEntryPointPackage(packageImport) {
				packageLabel = " (entry point)"
				packageIcon = "🚪"
			} else if isUtilityPackage(packageImport) {
				packageLabel = " (utility)"
				packageIcon = "🔧"
			}

			pkgMsg := fmt.Sprintf("%s %s%s", packageIcon, cleanPkg, packageLabel)
			indentedResult := clog.NewInfo(pkgMsg).WithIndent(1)
			clog.PrintColorMessage(indentedResult)
		}
	}
}

// DisplayPackageInfo shows detailed information about a specific package
func DisplayPackageInfo(workdir string, packageName string, ignoreTests bool) []clog.CheckResult {
	var results []clog.CheckResult

	// Display package name with a nice header
	headerMsg := "┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓"
	results = append(results, clog.NewInfo(headerMsg))

	titleMsg := fmt.Sprintf("┃  📦 PACKAGE DETAIL: %-36s ┃", packageName)
	results = append(results, clog.NewInfo(titleMsg))

	footerMsg := "┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛"
	results = append(results, clog.NewInfo(footerMsg))

	// Get package directory
	packagePath := workdir + strings.TrimPrefix(packageName, ModPath)
	packagePath = strings.Trim(packagePath, `"`)

	// Track if the package exists
	packageExists := false

	// Walk the file system to find package files
	filepath.Walk(packagePath, func(path string, info os.FileInfo, err error) error {
		// Log and skip if error is not nil
		if err != nil {
			results = append(results, clog.NewError(err.Error()))
			return nil
		}

		// Skip directories, non-go files and other invalid filenames
		if info.IsDir() || (len(info.Name()) > 3 && info.Name()[len(info.Name())-3:] != ".go") {
			return nil
		}

		dirString, fileString := filepath.Split(path)

		// Skip git files
		if strings.Contains(dirString, "/.git/") {
			return nil
		}

		// Skip test files if requested
		if ignoreTests && strings.HasSuffix(fileString, "_test.go") {
			return nil
		}

		packageExists = true

		// Get file imports
		fileImports, err := getImportsForFile(path)
		if err != nil {
			results = append(results, clog.NewError(err.Error()))
			return nil
		}

		// Display file information
		fileMsg := fmt.Sprintf("📄 File: %s", info.Name())
		results = append(results, clog.NewInfo(fileMsg))

		// Display imports with indentation
		if len(fileImports) > 0 {
			importMsg := "⤵️ Imports:"
			results = append(results, clog.NewInfo(importMsg).WithIndent(1))

			for _, fileImport := range fileImports {
				// Check if it's a project import and add special visual
				importIcon := "└─"
				if strings.Index(fileImport, ModPath) > 0 {
					importIcon = "🔗"
				}

				results = append(results,
					clog.NewInfo(fmt.Sprintf("%s %s", importIcon, fileImport)).WithIndent(2))
			}
		} else {
			results = append(results,
				clog.NewInfo("No imports").WithIndent(1))
		}

		// Add a small separator between files
		results = append(results, clog.NewInfo("┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄"))

		return nil
	})

	// If package doesn't exist, show error
	if !packageExists {
		results = append(results, clog.NewError(fmt.Sprintf("Package not found: %s", packageName)))
	}

	// Print all results
	for _, v := range results {
		clog.PrintColorMessage(v)
	}

	return results
}

// utilityPathPatterns is a list of path patterns that are considered utilities/shared libraries
// and should be placed at deeper levels
var utilityPathPatterns = []string{
	"utilities",
	"common",
	"pkg",
	"shared",
	"lib",
}

// isUtilityPackage determines if a package is a utility/shared library based on its path
func isUtilityPackage(packagePath string) bool {
	packagePath = strings.Trim(packagePath, `"`)
	for _, pattern := range utilityPathPatterns {
		if strings.Contains(packagePath, "/"+pattern+"/") {
			return true
		}
	}
	return false
}

// isEntryPointPackage determines if a package is an entry point (like main)
// Entry points are special cases that are allowed to import from any layer
func isEntryPointPackage(packagePath string) bool {
	packagePath = strings.Trim(packagePath, `"`)

	// Main packages are typically in the root or named 'cmd'
	if strings.HasSuffix(packagePath, "/.") ||
		strings.HasSuffix(packagePath, "/main") ||
		strings.Contains(packagePath, "/cmd/") {
		return true
	}

	return false
}

// getPackageDepth returns the logical depth of a package based on its path
// This helps ensure utility packages are placed at deeper levels
func getPackageDepth(packagePath string) int {
	packagePath = strings.Trim(packagePath, `"`)
	if isUtilityPackage(packagePath) {
		// Assign utility packages an extra level of depth
		return strings.Count(packagePath, "/") + 1
	}
	return strings.Count(packagePath, "/")
}

// SetUniqueLevels analyzes package dependencies and assigns unique levels
// taking into account both import relationships and directory structure
func SetUniqueLevels(packageMap map[string]PackageInfo) [][]string {
	var topLevelPackages []string
	packageLevels := make(map[string]int)

	// Find packages not imported by any other package (usually main)
	for _, packageInfo := range packageMap {
		packageIsMentionedInImports := false

		for _, packageDetails := range packageMap {
			if contains(packageDetails.Imports, packageInfo.Path) {
				packageIsMentionedInImports = true
			}
		}

		if !packageIsMentionedInImports {
			topLevelPackages = append(topLevelPackages, packageInfo.Path)
			packageLevels[packageInfo.Path] = 0
		}
	}

	// First phase: Assign initial levels based on import relationships
	packagesVisited := make(map[string]bool)
	for _, pkg := range topLevelPackages {
		packagesVisited[pkg] = true
	}

	// BFS traversal to assign levels
	var queue []string
	queue = append(queue, topLevelPackages...)

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		currentLevel := packageLevels[current]

		for _, imported := range packageMap[current].Imports {
			// If we haven't assigned a level or the new level would be deeper, update it
			if level, exists := packageLevels[imported]; !exists || level < currentLevel+1 {
				packageLevels[imported] = currentLevel + 1
			}

			if !packagesVisited[imported] {
				packagesVisited[imported] = true
				queue = append(queue, imported)
			}
		}
	}

	// Second phase: Adjust levels based on path structure
	for pkg := range packageLevels {
		pathDepth := getPackageDepth(pkg)
		if pathDepth > packageLevels[pkg] {
			// If a package's path suggests it should be deeper, increase its level
			packageLevels[pkg] = pathDepth
		}
	}

	// Find maximum level to determine array size
	maxLevel := 0
	for _, level := range packageLevels {
		if level > maxLevel {
			maxLevel = level
		}
	}

	// Organize packages by their computed levels
	packagesByLevel := make([][]string, maxLevel+1)
	for pkg, level := range packageLevels {
		packagesByLevel[level] = append(packagesByLevel[level], pkg)
	}

	// Remove empty levels
	result := make([][]string, 0)
	for _, pkgs := range packagesByLevel {
		if len(pkgs) > 0 {
			result = append(result, pkgs)
		}
	}

	return result
}

// Map builds a dependency map of the codebase
func Map(workdir string, ignoreTests bool) (map[string]PackageInfo, []clog.CheckResult) {
	var results []clog.CheckResult
	packageMap := make(map[string]PackageInfo)

	clog.Info("Analyzing packages...")

	// Walk through the project files
	filepath.Walk(workdir, func(path string, info os.FileInfo, err error) error {
		// Log and skip if error is not nil
		if err != nil {
			results = append(results, clog.NewError(err.Error()))
			return nil
		}

		// Skip directories, non go files and other invalid filenames
		if info.IsDir() || (len(info.Name()) > 3 && info.Name()[len(info.Name())-3:] != ".go") {
			return nil
		}

		dirString, fileString := filepath.Split(path)

		// Skip git files
		if strings.Contains(dirString, "/.git/") {
			return nil
		}

		// Skip test files if requested
		if ignoreTests && strings.HasSuffix(fileString, "_test.go") {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(workdir, path)
		if err != nil {
			results = append(results, clog.NewError(err.Error()))
			return nil
		}

		// Get file imports
		fileImports, err := getImportsForFile(path)
		if err != nil {
			results = append(results, clog.NewError(err.Error()))
			return nil
		}

		// Create package path
		packagePath := fmt.Sprintf("%q", ModPath+"/"+filepath.Dir(relPath))

		// Update or create package info
		if existingInfo, ok := packageMap[packagePath]; ok {
			// Update existing package
			existingInfo.Files = append(existingInfo.Files, fileString)

			// Add missing imports
			for _, packageImport := range fileImports {
				if strings.Index(packageImport, ModPath) > 0 {
					existingInfo.Imports = AppendStringIfMissing(existingInfo.Imports, packageImport)
				}
			}

			packageMap[packagePath] = existingInfo
		} else {
			// Create new package
			packageInfo := PackageInfo{
				Path:  packagePath,
				Files: []string{fileString},
				Level: 0,
			}

			// Add imports from project
			var packageImports []string
			for _, packageImport := range fileImports {
				if strings.Index(packageImport, ModPath) > 0 {
					packageImports = AppendStringIfMissing(packageImports, packageImport)
				}
			}

			packageInfo.Imports = packageImports
			packageMap[packagePath] = packageInfo
		}

		return nil
	})

	// Print any errors that occurred during mapping
	for _, v := range results {
		clog.PrintColorMessage(v)
	}

	return packageMap, results
}

// AppendStringIfMissing adds a string to a slice if it's not already present
func AppendStringIfMissing(slice []string, item string) []string {
	for _, element := range slice {
		if element == item {
			return slice
		}
	}
	return append(slice, item)
}

// getImportsForFile parses a Go file and returns its imports
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
