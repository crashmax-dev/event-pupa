package loggerImplementation

import "runtime"

func getOSFilePath(filePath string) string {
	if runtime.GOOS == "windows" {
		return "winfile:///" + filePath
	} else {
		return filePath
	}
}
