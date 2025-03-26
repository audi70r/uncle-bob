package clog

const (
	// Color constants
	Teal   color = "\033[1;36m"
	Green  color = "\033[1;32m"
	Yellow color = "\033[1;33m"
	Purple color = "\033[1;35m"
	Red    color = "\033[1;31m"
	Blue   color = "\033[1;34m"
	Reset  color = "\033[0m"

	// Default color assignments
	teal   = Teal
	green  = Green
	yellow = Yellow
	purple = Purple
	red    = Red
	reset  = Reset

	// Result types
	ResultErr     resultType = "ERROR"
	ResultInfo    resultType = "INFO"
	ResultWarning resultType = "WARNING"
	ResultSuccess resultType = "SUCCESS"

	// Legacy constants (for backward compatibility)
	resultErr     = ResultErr
	resultInfo    = ResultInfo
	resultWarning = ResultWarning
)
