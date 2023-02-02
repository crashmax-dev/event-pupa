package loggerImplementation

import "runtime"

func getOSFilePath(filePath string) string {
	if //goland:noinspection GoBoolExpressions
	runtime.GOOS == "windows" {
		return "winfile:///" + filePath
	} else {
		return filePath
	}
}
