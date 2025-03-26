package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/audi70r/uncle-bob/checker"
	"github.com/audi70r/uncle-bob/utilities/clog"
)

const (
	AppVersion = "1.2.0"
	AppAuthor  = "dmitri@nuage.ee"
)

// printLogo prints the application logo
func printLogo() {
	logo := []string{
		` /\ /\ _ __   ___| | ___    / __\ ___ | |__  `,
		`/ / \ \ '_ \ / __| |/ _ \  /__\/// _ \| '_ \ `,
		`\ \_/ / | | | (__| |  __/ / \/  \ (_) | |_) |`,
		` \___/|_| |_|\___|_|\___| \_____/\___/|_.__/ `,
		fmt.Sprintf(`v%s                         %s `, AppVersion, AppAuthor),
	}

	fmt.Println(string(clog.Blue) + strings.Join(logo, "\n") + string(clog.Reset))
	fmt.Println("")
}

func main() {
	// Define command line flags
	fileImports := flag.String("package-imports", "", "Show detailed information about package imports")
	strictFlag := flag.Bool("strict", false, "Do strict checking, allow only one level inward imports")
	ignoreTests := flag.Bool("ignore-tests", false, "Ignore imports of test files")
	pathFlag := flag.String("path", "", "Specify the path to analyze (default: current directory)")
	quietFlag := flag.Bool("quiet", false, "Only show warnings and errors")
	silentFlag := flag.Bool("silent", false, "Only show errors")
	noColorFlag := flag.Bool("no-color", false, "Disable colored output")
	versionFlag := flag.Bool("version", false, "Show version information")

	// Parse command line flags
	flag.Parse()

	// Handle version flag
	if *versionFlag {
		fmt.Printf("Uncle Bob %s\n", AppVersion)
		return
	}

	// Add custom help message
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help") {
		fmt.Println("Uncle Bob - Clean Architecture Dependency Checker")
		fmt.Println("\nUsage:")
		fmt.Println("  uncle-bob [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  uncle-bob                           # Check current directory")
		fmt.Println("  uncle-bob -path=/path/to/project    # Check specific directory")
		fmt.Println("  uncle-bob -strict                   # Use strict mode (only allows one level inward imports)")
		fmt.Println("  uncle-bob -package-imports=\"pkg/foo\" # Show imports for a specific package")
		return
	}

	// Configure logging
	if *silentFlag {
		clog.SetLogLevel(clog.LevelError)
	} else if *quietFlag {
		clog.SetLogLevel(clog.LevelWarning)
	}

	// Configure color output
	if *noColorFlag {
		clog.DisableColor()
	}

	// Show the logo unless silent mode is enabled
	if !*silentFlag {
		printLogo()
	}

	// Determine working directory (either specified path or current directory)
	var workDir string
	var err error

	if *pathFlag != "" {
		// Use the specified path
		workDir, err = filepath.Abs(*pathFlag)
		if err != nil {
			clog.Error(fmt.Sprintf("Failed to resolve specified path: %s", err.Error()))
			os.Exit(1)
		}

		// Verify the path exists
		fileInfo, err := os.Stat(workDir)
		if err != nil {
			clog.Error(fmt.Sprintf("Path does not exist or is not accessible: %s", err.Error()))
			os.Exit(1)
		}

		if !fileInfo.IsDir() {
			clog.Error("Specified path is not a directory")
			os.Exit(1)
		}
	} else {
		// Use current working directory
		workDir, err = os.Getwd()
		if err != nil {
			clog.Error(fmt.Sprintf("Failed to get working directory: %s", err.Error()))
			os.Exit(1)
		}
	}

	// Locate go.mod file
	checker.LocateGoMod(workDir)

	// Handle package-imports flag
	if *fileImports != "" {
		results := checker.DisplayPackageInfo(workDir, *fileImports, *ignoreTests)

		// Check if any errors occurred
		for _, res := range results {
			if res.Type == clog.ResultErr {
				os.Exit(1)
			}
		}
		return
	}

	// Generate package map
	packageMap, results := checker.Map(workDir, *ignoreTests)

	// Check if any errors occurred during mapping
	for _, res := range results {
		if res.Type == clog.ResultErr {
			os.Exit(1)
		}
	}

	// Generate package levels
	packageLevels := checker.SetUniqueLevels(packageMap)

	// Display package level information
	checker.LevelsInfo(packageLevels)

	// Check for violations
	results = checker.CheckLevels(packageMap, packageLevels, *strictFlag)

	// Output final result
	if checker.HasViolations(results) {
		clog.Error("Issues detected, Uncle Bob is Sad :(")
		clog.Info("\nReview the suggestions above to fix architectural issues.")
		clog.Info("For Clean Architecture principles, visit: https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html")
		os.Exit(1)
	} else {
		clog.Info("Well done, Uncle Bob is Proud :)")
	}
}
