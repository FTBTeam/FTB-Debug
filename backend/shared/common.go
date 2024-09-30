package shared

import "os"

var (
	GitCommit string
	Version   string
)

func DoesPathExist(filePath string) bool {
	if _, err := os.Stat(filePath); err == nil {
		return true
	}
	return false
}
