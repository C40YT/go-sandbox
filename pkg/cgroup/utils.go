package cgroup

import (
	"io/ioutil"
	"os"
	"path"
)

// ensureDirExists creates directories if the path not exists
func ensureDirExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, dirPerm)
	}
	return nil
}

// createTempDir creates path for sub-cgroup with given group and prefix
func createTempDir(elem ...string) (string, error) {
	base := path.Join(elem...)
	err := ensureDirExists(base)
	if err != nil {
		return "", err
	}
	return ioutil.TempDir(base, "")
}

func remove(filename string) (err error) {
	if filename != "" {
		err = os.Remove(filename)
	}
	return
}

func removeAll(filenames ...string) (err error) {
	for _, filename := range filenames {
		e1 := remove(filename)
		if e1 != nil && err == nil {
			err = e1
		}
	}
	return
}
