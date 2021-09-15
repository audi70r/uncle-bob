package main

import (
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

	workDir, wrkDirErr := os.Getwd()
	if wrkDirErr != nil {
		log.Println(wrkDirErr)
	}

	checker.LocateGoMod(workDir)
	packageMap, mapResults := checker.Map(workDir)

	packageLevels := checker.SetLevels(packageMap)

	levelCheckResults := checker.CheckLevels(packageMap, packageLevels)

	for _, v := range mapResults {
		clog.PrintColorMessage(v)
	}

	for _, v := range levelCheckResults {
		clog.PrintColorMessage(v)
	}
}
