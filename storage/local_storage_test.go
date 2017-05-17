package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalStorage(t *testing.T) {
	assert := assert.New(t)

	root, err := ioutil.TempDir("", "tri_local_test_")
	if !assert.NoError(err, "failed to create temp directory") {
		t.FailNow()
	}

	cleanupTestPath := func() error {
		// Remove previous temp files
		files, err := ioutil.ReadDir(root)
		if err != nil {
			return err
		}
		for _, f := range files {
			os.Remove(filepath.Join(root, f.Name()))
		}

		/*
			/
			/folder_a/file_a
			/folder_a/folder_b/file_b
			/folder_a/folder_empty
			/folder_empty
		*/
		err = os.MkdirAll(filepath.Join(root, "folder_a", "folder_b"), 0744)
		if err != nil {
			return err
		}
		err = os.MkdirAll(filepath.Join(root, "folder_a", "folder_empty"), 0744)
		if err != nil {
			return err
		}
		err = os.MkdirAll(filepath.Join(root, "folder_empty"), 0744)
		if err != nil {
			return err
		}
		var f *os.File
		f, err = os.OpenFile(filepath.Join(root, "file_a"), os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			return err
		}
		f.Close()
		f2, err2 := os.OpenFile(filepath.Join(root, "folder_a", "folder_b", "file_b"), os.O_RDWR|os.O_CREATE, 0644)
		if err2 != nil {
			return err2
		}
		f2.Close()
		return nil
	}

	err = cleanupTestPath()
	if !assert.NoError(err, "failed to init directory dir") {
		t.FailNow()
	}

	localStorage, err := NewLocalStorage(root)
	if !assert.NoError(err, "failed to create local storage") {
		t.FailNow()
	}

	testStorage := NewStorageTester(localStorage, cleanupTestPath)
	t.Run("LocalStorage", RunStorageTests(testStorage))
}
