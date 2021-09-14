package clog

const (
	teal   color = "\033[1;36m"
	green        = "\033[1;32m"
	yellow       = "\033[1;33m"
	purple       = "\033[1;35m"
	red          = "\033[1;31m"
	reset        = "\033[0m"

	resultErr     resultType = "ERROR"
	resultInfo    resultType = "INFO"
	resultWarning resultType = "WARNING"
)
