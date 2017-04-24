package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

var (
	testStorageSource *LocalStorage
)

func init() {
	root, err := setupLocalStorage()
	if err != nil {
		panic(err)
	}
	testStorageSource, err = NewLocalStorage(root)
	if err != nil {
		panic(err)
	}
}

// Setup a dummy local storage with some files and recursion
// Returns the root path of the temporary directory
func setupLocalStorage() (string, error) {
	root, err := ioutil.TempDir("", "tri_local_test_")
	if err != nil {
		return root, err
	}

	// Remove previous temp files
	files, err := ioutil.ReadDir(root)
	if err != nil {
		return root, err
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
	err = os.MkdirAll(filepath.Join(root, "folder_a", "folder_b"), 0644)
	if err != nil {
		return root, err
	}
	err = os.MkdirAll(filepath.Join(root, "folder_a", "folder_empty"), 0644)
	if err != nil {
		return root, err
	}
	err = os.MkdirAll(filepath.Join(root, "folder_empty"), 0644)
	if err != nil {
		return root, err
	}
	var f *os.File
	f, err = os.OpenFile(filepath.Join(root, "file_a"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return root, err
	}
	defer f.Close()
	f2, err2 := os.OpenFile(filepath.Join(root, "folder_a", "folder_b", "file_b"), os.O_RDWR|os.O_CREATE, 0644)
	if err2 != nil {
		return root, err2
	}
	defer f2.Close()
	return root, nil
}

func TestNoEscape(t *testing.T) {
	t.Log("Temp dir:", testStorageSource.Root)
	_, err := testStorageSource.List("..")
	if err != ErrNotInRoot {
		t.Fatalf("We should have been chrooted, got error instead: %s", err)
	}
}

func TestDownloadUpload(t *testing.T) {
	exempleText := []byte("hello")
	name := "test_upload_download"
	f, err := testStorageSource.Upload(name)
	if err != nil {
		t.Fatalf("Failed to upload: %s", err)
	}
	f.Write(exempleText)
	f.Close()
	defer os.Remove(filepath.Join(testStorageSource.Root, name))

	d, err2 := testStorageSource.Download(name)
	if err2 != nil {
		t.Fatalf("Failed to upload: %s", err2)
	}
	defer d.Close()
	text, err3 := ioutil.ReadAll(d)
	if err3 != nil {
		t.Fatalf("Failed to upload: %s", err3)
	}
	if string(text) != string(exempleText) {
		t.Fatalf("Text mismatch %s != %s", exempleText, text)
	}
}

func TestList(t *testing.T) {
	listing, err := testStorageSource.List(".")
	if err != nil {
		t.Fatalf("Listing error: %s", err)
	}
	if len(listing) != 3 { // Should be 2 folders + 1 file
		t.Fatal("Listing should contain 2 folders and 1 file:", listing)
	}
	sort.Sort(sortAlphabetical(listing))
	if listing[0].Directory || listing[0].Name != "file_a" {
		t.Fatalf("listing[0] is incorrect: %v", listing[0])
	}
	if !listing[1].Directory || listing[1].Name != "folder_a" {
		t.Fatalf("listing[1] is incorrect: %v", listing[1])
	}
	if !listing[2].Directory || listing[2].Name != "folder_empty" {
		t.Fatalf("listing[2] is incorrect: %v", listing[2])
	}

	listing, err = testStorageSource.List("folder_empty")
	if err != nil {
		t.Fatalf("Listing error: %s", err)
	}
	if len(listing) != 0 {
		t.Fatal("Listing of empty folder should be empty")
	}
}

func TestMkdir(t *testing.T) {
	dirName := "test_mkdir"
	err := testStorageSource.Mkdir(dirName)
	if err != nil {
		t.Fatalf("Failed to make dir: %s", err)
	}
	defer os.Remove(filepath.Join(testStorageSource.Root, dirName))
	listing, _ := testStorageSource.List(".")
	for _, l := range listing {
		if l.Name == dirName && l.Directory {
			return
		}
	}
	t.Fatalf("Failed to find created directory: %s", listing)
}

func TestMove(t *testing.T) {
	oldName := "test_move"
	newName := "test_move2"
	err := ioutil.WriteFile(filepath.Join(testStorageSource.Root, oldName), []byte("hello"), defaultPerms)
	if err != nil {
		t.Fatalf("Failed to create test file: %s", err)
	}
	err = testStorageSource.Move(oldName, newName)
	if err != nil {
		t.Fatalf("Failed to move file: %s", err)
	}
	defer os.Remove(filepath.Join(testStorageSource.Root, newName))
	// Check that it was acutally moved
	found := false
	listing, _ := testStorageSource.List(".")
	for _, l := range listing {
		if l.Name == oldName {
			t.Fatal("Found old file still there")
		}
		if l.Name == newName {
			found = true
		}
	}
	if !found {
		t.Fatalf("Failed to find new file: %s", listing)
	}
}

func TestRemove(t *testing.T) {
	// Create temp file
	name := "test_remove"
	err := ioutil.WriteFile(filepath.Join(testStorageSource.Root, name), []byte("hello"), defaultPerms)
	if err != nil {
		t.Fatalf("Failed to create test file: %s", err)
	}

	testStorageSource.Remove(name)
	listing, _ := testStorageSource.List(".")
	for _, l := range listing {
		if l.Name == name {
			t.Fatal("Previous file was not removed")
		}
	}
}
