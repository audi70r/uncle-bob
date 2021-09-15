package main

import (
	"flag"
	"fmt"
	"github.com/audi70r/go-archangel/checker"
	"github.com/audi70r/go-archangel/utilities/clog"
	"log"
	"os"
)

func PrintAA() {
	aa := []string{
		`   _____                        _                            _ `,
		`  / ____|        /\            | |                          | |`,
		` | |  __  ___   /  \   _ __ ___| |__   __ _ _ __   __ _  ___| |`,
		` | | |_ |/ _ \ / /\ \ | '__/ __| '_ \ / _  | '_ \ / _  |/ _ \ |`,
		` | |__| | (_) / ____ \| | | (__| | | | (_| | | | | (_| |  __/ |`,
		`  \_____|\___/_/    \_\_|  \___|_| |_|\__,_|_| |_|\__, |\___|_|`,
		`                                                   __/ |       `,
		`                                                  |___/        `,
		`v1.0                                      by Dmitri Beltjukov  `,
	}
	for _, s := range aa {
		fmt.Println(s)
	}
	fmt.Println("")
}

func main() {
	PrintAA()

	fileImports := flag.String("package-imports", "", "show detailed information about package imports")
	strictFlag := flag.Bool("strict", false, "do strict checking, do not allow same level imports")
	ignoreTests := flag.Bool("ignore-tests", false, "ignore imports of test files")

	flag.Parse()

	workDir, wrkDirErr := os.Getwd()
	if wrkDirErr != nil {
		log.Println(wrkDirErr)
	}

	checker.LocateGoMod(workDir)

	if *fileImports != "" {
		clog.Info("Package: " + *fileImports)
		packageimportsInfo := checker.DisplayPackageInfo(workDir, *fileImports, *ignoreTests)

		for _, v := range packageimportsInfo {
			clog.PrintColorMessage(v)
		}

		return
	}

	packageMap, mapResults := checker.Map(workDir, *ignoreTests)

	packageLevels := checker.SetUniqueLevels(packageMap)

	levelsInfo := checker.LevelsInfo(packageLevels)

	levelCheckResults := checker.CheckLevels(packageMap, packageLevels, *strictFlag)

	for _, v := range levelsInfo {
		clog.PrintColorMessage(v)
	}

	for _, v := range mapResults {
		clog.PrintColorMessage(v)
	}

	for _, v := range levelCheckResults {
		clog.PrintColorMessage(v)
	}
}
