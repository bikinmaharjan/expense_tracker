package utils

import (
	"os"
	"strconv"
)

// DeleteFile deletes a file from the file system
func DeleteFile(path string) error {
	return os.Remove(path)
}

// ParseInt parses a string to an integer with a default value
func ParseInt(s string, defaultValue int) (int, error) {
	if s == "" {
		return defaultValue, nil
	}
	return strconv.Atoi(s)
}
