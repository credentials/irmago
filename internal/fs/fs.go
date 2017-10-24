package fs

import (
	"crypto/rand"
	"encoding/hex"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
)

// AssertPathExists returns nil only if it has been successfully
// verified that the specified path exists.
func AssertPathExists(path string) error {
	exist, err := PathExists(path)
	if err != nil {
		return err
	}
	if !exist {
		return errors.Errorf("Path %s does not exist", path)
	}
	return nil
}

// PathExists checks if the specified path exists.
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func EnsureDirectoryExists(path string) error {
	exists, err := PathExists(path)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return os.Mkdir(path, 0700)
}

// Save the filecontents at the specified path atomically:
// - first save the content in a temp file with a random filename in the same dir
// - then rename the temp file to the specified filepath, overwriting the old file
func SaveFile(filepath string, content []byte) (err error) {
	dir := path.Dir(filepath)

	// Read random data for filename and convert to hex
	randBytes := make([]byte, 16)
	_, err = rand.Read(randBytes)
	if err != nil {
		return
	}
	tempfilename := hex.EncodeToString(randBytes)

	// Create temp file
	err = ioutil.WriteFile(dir+"/"+tempfilename, content, 0600)
	if err != nil {
		return
	}

	// Rename, overwriting old file
	return os.Rename(dir+"/"+tempfilename, filepath)
}