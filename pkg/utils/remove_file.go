package utils

import (
	"fmt"
	"os"
)


func RemoveFile(filePath string) error {
	// Remove the file
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("error removing file %s: %v", filePath, err)
	}
	return nil
}
