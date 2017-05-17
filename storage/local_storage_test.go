package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var (
	testRoot          string
	testStorageSource *LocalStorage
)

func init() {
	err := setupLocalStorage()
	if err != nil {
		panic(err)
	}
	testStorageSource, err = NewLocalStorage(testRoot)
	if err != nil {
		panic(err)
	}
}

// Setup a dummy local storage with some files and recursion
// Returns the root path of the temporary directory
func setupLocalStorage() error {
	var err error
	testRoot, err = ioutil.TempDir("", "tri_local_test_")
	if err != nil {
		return err
	}
	return cleanupTestPath()
}

func cleanupTestPath() error {
	// Remove previous temp files
	files, err := ioutil.ReadDir(testRoot)
	if err != nil {
		return err
	}
	for _, f := range files {
		os.Remove(filepath.Join(testRoot, f.Name()))
	}

	/*
		/
		/folder_a/file_a
		/folder_a/folder_b/file_b
		/folder_a/folder_empty
		/folder_empty
	*/
	err = os.MkdirAll(filepath.Join(testRoot, "folder_a", "folder_b"), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(testRoot, "folder_a", "folder_empty"), 0744)
	if err != nil {
		return err
	}
	err = os.MkdirAll(filepath.Join(testRoot, "folder_empty"), 0744)
	if err != nil {
		return err
	}
	var f *os.File
	f, err = os.OpenFile(filepath.Join(testRoot, "file_a"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	f.Close()
	f2, err2 := os.OpenFile(filepath.Join(testRoot, "folder_a", "folder_b", "file_b"), os.O_RDWR|os.O_CREATE, 0644)
	if err2 != nil {
		return err2
	}
	f2.Close()
	return nil
}

// Make we can't escape root
func TestNoEscape(t *testing.T) {
	_, err := testStorageSource.List("..")
	if errors.Cause(err) != ErrNotInRoot {
		t.Fatalf("We should have been chrooted, got error instead: %s", err)
	}
}

func TestDownloadUpload(t *testing.T) {
	assert := assert.New(t)
	defer cleanupTestPath()
	exampleText := []byte("hello")
	name := "test_upload_download"
	modTime := time.Date(2017, time.May, 17, 20, 10, 6, 0, time.UTC)

	// Upload test
	f, err := testStorageSource.Upload(name, modTime)
	assert.NoError(err, "failed to open for upload")
	defer f.Close()
	_, err = f.Write(exampleText)
	assert.NoError(err, "failed to write")
	err = f.Close()
	assert.NoError(err, "failed to close")

	// Test Download
	d, err := testStorageSource.Download(name)
	assert.NoError(err, "Failed to open for download")
	defer d.Close()
	text, err := ioutil.ReadAll(d)
	assert.NoError(err, "Failed to read download file")
	err = d.Close()
	assert.NoError(err, "Failed to close download file")

	assert.Equal(string(text), string(exampleText))
}

func TestUploadChtime(t *testing.T) {
	assert := assert.New(t)
	defer cleanupTestPath()
	exampleText := []byte("hello")
	name := "test_upload_download"
	modTime := time.Date(2017, time.May, 17, 20, 10, 6, 0, time.UTC)

	// Upload test
	f, err := testStorageSource.Upload(name, modTime)
	assert.NoError(err, "failed to open for upload")
	defer f.Close()
	_, err = f.Write(exampleText)
	assert.NoError(err, "failed to write")
	err = f.Close()
	assert.NoError(err, "failed to close")

	// List folder to check
	listing, err := testStorageSource.List(".")
	assert.NoError(err, "failed to list root")
	for _, l := range listing {
		if l.Name == name {
			assert.Equal(modTime.Second(), l.Modified.Second())
			return // Found the file
		}
	}
	t.Fatal("Could not find our uploaded file: ", listing)

}

func TestList(t *testing.T) {
	assert := assert.New(t)

	// List root
	listing, err := testStorageSource.List(".")
	assert.NoError(err, "failed to list root")
	assert.Equal(3, len(listing), "listing should contain 2 folders and 1 file")

	sort.Sort(sortAlphabetical(listing))
	assert.Equal(false, listing[0].IsDirectory)
	assert.Equal("file_a", listing[0].Name)

	assert.Equal(true, listing[1].IsDirectory)
	assert.Equal("folder_a", listing[1].Name)

	assert.Equal(true, listing[2].IsDirectory)
	assert.Equal("folder_empty", listing[2].Name)

	// List empty directory
	listing, err = testStorageSource.List("folder_empty")
	assert.NoError(err, "failed to list empty folder")
	assert.Equal(0, len(listing))
}

func TestMkdir(t *testing.T) {
	defer cleanupTestPath()
	assert := assert.New(t)

	dirName := "test_mkdir"
	err := testStorageSource.Mkdir(dirName)
	assert.NoError(err, "failed to mkdir")

	listing, err := testStorageSource.List(".")
	assert.NoError(err, "failed to list root")
	for _, l := range listing {
		if l.Name == dirName && l.IsDirectory {
			return
		}
	}
	t.Fatalf("Failed to find created directory: %s", listing)
}

func TestMove(t *testing.T) {
	assert := assert.New(t)
	defer cleanupTestPath()

	testingMove := func(src, dst string) {
		err := testStorageSource.Move(src, dst)
		assert.NoError(err, "failed to move")

		// Check if we moved
		found := false
		listing, err := testStorageSource.List(".")
		assert.NoError(err, "failed to list")
		for _, l := range listing {
			assert.NotEqual(src, l.Name, "found old file still there")
			if l.Name == dst {
				found = true
			}
		}
		assert.True(found, "new folder was not found", listing)
	}

	// Can't move a folder for now
	/*
		// Test moving a folder
		dirSrcName := "test_move_dir_src"
		dirDstName := "test_move_dir_dst"
		err := testStorageSource.Mkdir(dirSrcName)
		assert.NoError(err, "failed to mkdir")
		testingMove(dirSrcName, dirDstName)
	*/

	// Test moving a file
	fileSrcName := "test_move_file_src"
	fileDstName := "test_move_file_dst"
	f, err := testStorageSource.Upload(fileSrcName, time.Now())
	assert.NoError(err, "failed to open for upload")
	err = f.Close()
	assert.NoError(err, "failed to close upload")
	testingMove(fileSrcName, fileDstName)
}

func TestRemove(t *testing.T) {
	assert := assert.New(t)

	testingRemove := func(name string) {
		err := testStorageSource.Remove(name)
		assert.NoError(err, "failed to remove")
		// Make sure we actually deleted the file
		listing, _ := testStorageSource.List(".")
		for _, l := range listing {
			assert.NotEqual(name, l.Name, "previous file was not removed")
		}
	}

	// Test removing a file
	fileName := "test_remove_file"
	f, err := testStorageSource.Upload(fileName, time.Now())
	assert.NoError(err, "failed to open for upload")
	err = f.Close()
	assert.NoError(err, "failed to close upload")
	testingRemove(fileName)

	// Test removing a folder
	folderName := "test_remove_folder"
	err = testStorageSource.Mkdir(folderName)
	assert.NoError(err, "failed to create folder")
	testingRemove(folderName)
}
