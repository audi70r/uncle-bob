package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/audi70r/go-archangel/checker"
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
		_ = checker.DisplayPackageInfo(workDir, *fileImports, *ignoreTests)

		return
	}

	packageMap, _ := checker.Map(workDir, *ignoreTests)

	packageLevels := checker.SetUniqueLevels(packageMap)

	_ = checker.LevelsInfo(packageLevels)

	_ = checker.CheckLevels(packageMap, packageLevels, *strictFlag)
}
