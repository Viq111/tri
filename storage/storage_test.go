package storage

import (
	"io/ioutil"
	"sort"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

/*
StorageTester is the interface that wraps method to be able to run the tests
You should create an object that satify this interface and you can then run
the full test battery by calling RunStorageTests(t *testing.T)

The initial file tree should look like this:
        /
        /folder_a/file_a
        /folder_a/folder_b/file_b
        /folder_a/folder_empty
        /folder_empty
*/
type StorageTester interface {
	// Cleanup reset the file directory to the initial state
	Cleanup() error
	// GetStorage returns an object that should implement the Storage interface
	GetStorage() Storage
}

type STester struct {
	cleanup func() error
	s       Storage
}

func NewStorageTester(s Storage, cleanup func() error) *STester {
	return &STester{
		cleanup: cleanup,
		s:       s,
	}
}
func (s *STester) Cleanup() error {
	return s.cleanup()
}
func (s *STester) GetStorage() Storage {
	return s.s
}

func RunStorageTests(st StorageTester) func(t *testing.T) {
	s := st.GetStorage()

	// underlying tests should take a *testing.T and Storage
	// and returns if they need cleanup after
	testFunc := func(f func(*testing.T, Storage) bool) func(*testing.T) {
		return func(t2 *testing.T) {
			needCleanup := f(t2, s)
			if needCleanup {
				st.Cleanup()
			}
		}
	}

	return func(t *testing.T) {
		// Run the tests
		t.Run("TestStorageNoEscape", testFunc(testStorageNoEscape))
		t.Run("TestStorageDownloadUpload", testFunc(testStorageDownloadUpload))
		t.Run("TestStorageUploadChtime", testFunc(testStorageUploadChtime))
		t.Run("TestStorageList", testFunc(testStorageList))
		t.Run("TestStorageMkdir", testFunc(testStorageMkdir))
		t.Run("TestStorageMove", testFunc(testStorageMove))
		t.Run("TestStorageRemove", testFunc(testStorageRemove))
	}
}

// Make we can't escape root
func testStorageNoEscape(t *testing.T, s Storage) (needCleanup bool) {
	_, err := s.List("..")
	if errors.Cause(err) != ErrNotInRoot {
		t.Fatalf("We should have been chrooted, got error instead: %s", err)
	}
	return false
}

func testStorageDownloadUpload(t *testing.T, s Storage) (needCleanup bool) {
	assert := assert.New(t)
	needCleanup = true
	exampleText := []byte("hello")
	name := "test_upload_download"
	modTime := time.Date(2017, time.May, 17, 20, 10, 6, 0, time.UTC)

	// Upload test
	f, err := s.Upload(name, modTime)
	assert.NoError(err, "failed to open for upload")
	defer f.Close()
	_, err = f.Write(exampleText)
	assert.NoError(err, "failed to write")
	err = f.Close()
	assert.NoError(err, "failed to close")

	// Test Download
	d, err := s.Download(name)
	assert.NoError(err, "Failed to open for download")
	defer d.Close()
	text, err := ioutil.ReadAll(d)
	assert.NoError(err, "Failed to read download file")
	err = d.Close()
	assert.NoError(err, "Failed to close download file")

	assert.Equal(string(text), string(exampleText))
	return
}

func testStorageUploadChtime(t *testing.T, s Storage) (needCleanup bool) {
	assert := assert.New(t)
	needCleanup = true
	exampleText := []byte("hello")
	name := "test_upload_download"
	modTime := time.Date(2017, time.May, 17, 20, 10, 6, 0, time.UTC)

	// Upload test
	f, err := s.Upload(name, modTime)
	assert.NoError(err, "failed to open for upload")
	defer f.Close()
	_, err = f.Write(exampleText)
	assert.NoError(err, "failed to write")
	err = f.Close()
	assert.NoError(err, "failed to close")

	// List folder to check
	listing, err := s.List(".")
	assert.NoError(err, "failed to list root")
	for _, l := range listing {
		if l.Name == name {
			assert.Equal(modTime.Second(), l.Modified.Second())
			return // Found the file
		}
	}
	t.Fatal("Could not find our uploaded file: ", listing)
	return
}

func testStorageList(t *testing.T, s Storage) (needCleanup bool) {
	assert := assert.New(t)

	// List root
	listing, err := s.List(".")
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
	listing, err = s.List("folder_empty")
	assert.NoError(err, "failed to list empty folder")
	assert.Equal(0, len(listing))
	return
}

func testStorageMkdir(t *testing.T, s Storage) (needCleanup bool) {
	assert := assert.New(t)
	needCleanup = true

	dirName := "test_mkdir"
	err := s.Mkdir(dirName)
	assert.NoError(err, "failed to mkdir")

	listing, err := s.List(".")
	assert.NoError(err, "failed to list root")
	for _, l := range listing {
		if l.Name == dirName && l.IsDirectory {
			return
		}
	}
	t.Fatalf("Failed to find created directory: %s", listing)
	return
}

func testStorageMove(t *testing.T, s Storage) (needCleanup bool) {
	assert := assert.New(t)
	needCleanup = true

	testingMove := func(src, dst string) {
		err := s.Move(src, dst)
		assert.NoError(err, "failed to move")

		// Check if we moved
		found := false
		listing, err := s.List(".")
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
	f, err := s.Upload(fileSrcName, time.Now())
	assert.NoError(err, "failed to open for upload")
	err = f.Close()
	assert.NoError(err, "failed to close upload")
	testingMove(fileSrcName, fileDstName)
	return
}

func testStorageRemove(t *testing.T, s Storage) (needCleanup bool) {
	assert := assert.New(t)
	needCleanup = true

	testingRemove := func(name string) {
		err := s.Remove(name)
		assert.NoError(err, "failed to remove")
		// Make sure we actually deleted the file
		listing, _ := s.List(".")
		for _, l := range listing {
			assert.NotEqual(name, l.Name, "previous file was not removed")
		}
	}

	// Test removing a file
	fileName := "test_remove_file"
	f, err := s.Upload(fileName, time.Now())
	assert.NoError(err, "failed to open for upload")
	err = f.Close()
	assert.NoError(err, "failed to close upload")
	testingRemove(fileName)

	// Test removing a folder
	folderName := "test_remove_folder"
	err = s.Mkdir(folderName)
	assert.NoError(err, "failed to create folder")
	testingRemove(folderName)
	return
}
