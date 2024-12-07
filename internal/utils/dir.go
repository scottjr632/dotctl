package utils

import "os"

func IsDirectoryEmpty(dirPath string) (bool, error) {
	dir, err := os.Open(dirPath)
	if err != nil {
		return false, err
	}
	defer dir.Close()

	// Read the contents of the directory
	contents, err := dir.Readdirnames(0) // 0 to read all entries
	if err != nil {
		return false, err
	}

	// Check if the contents slice is empty
	return len(contents) == 0, nil
}
