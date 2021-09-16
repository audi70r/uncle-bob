package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/audi70r/uncle-bob/checker"
)

func PrintAA() {
	aa := []string{
		` /\ /\ _ __   ___| | ___    / __\ ___ | |__  `,
		`/ / \ \ '_ \ / __| |/ _ \  /__\/// _ \| '_ \ `,
		`\ \_/ / | | | (__| |  __/ / \/  \ (_) | |_) |`,
		` \___/|_| |_|\___|_|\___| \_____/\___/|_.__/ `,
		`v1.0                         dmitri@nuage.ee `,
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

	checker.LevelsInfo(packageLevels)

	checker.CheckLevels(packageMap, packageLevels, *strictFlag)

	if checker.UncleBobIsSad {
		fmt.Println("Issues detected, Uncle Bob is Sad :(")
		os.Exit(1)
	}

	fmt.Println("Well done, Uncle Bob is Proud :)")
}
