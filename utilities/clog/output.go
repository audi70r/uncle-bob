package clog

import "fmt"

type resultType string
type color string

// CheckResult is the result of dependency checking
type CheckResult struct {
	resultType resultType
	Message    string
	color      color
}

func Error(msg string) {
	res := NewError(msg)

	PrintColorMessage(res)
}

func Info(msg string) {
	res := NewInfo(msg)

	PrintColorMessage(res)
}

func Warning(msg string) {
	res := NewWarning(msg)

	PrintColorMessage(res)
}

func NewError(msg string) CheckResult {
	return CheckResult{
		resultType: resultErr,
		Message:    msg,
		color:      red,
	}
}

func NewInfo(msg string) CheckResult {
	return CheckResult{
		resultType: resultInfo,
		Message:    msg,
		color:      teal,
	}
}

func NewWarning(msg string) CheckResult {
	return CheckResult{
		resultType: resultWarning,
		Message:    msg,
		color:      yellow,
	}
}

func PrintColorMessage(cr CheckResult) {
	fmt.Printf("%s%-11s%s\n%s", cr.color, "["+cr.resultType+"]", cr.Message, reset)
}
