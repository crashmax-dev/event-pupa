package logger

import "runtime"

func getOSFilePath(filePath string) string {
	os := runtime.GOOS
	if os == "windows" {
		return "winfile:///" + filePath
	} else {
		return filePath
	}
}
