package storage

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	defaultPerms          = 0660
	defaultDirectoryPerms = 0770
)

// Defines exported errors
var (
	ErrAlreadyExist = errors.New("path already exists")
	ErrDirectory    = errors.New("path is a directory")
	ErrNotInRoot    = errors.New("path is not in the root of the given storage")
	ErrNotExist     = errors.New("path does not exist")
)

// LocalStorage implements Storage for local to a root path
// It is not safely chrooted to root directory, you should chroot the process
// if you want more security
type LocalStorage struct {
	Root string // Root is an absolute path
}

// NewLocalStorage returns a new local storage at root.
// It will check if the directory is writtable
func NewLocalStorage(root string) (l *LocalStorage, err error) {
	root, err = filepath.Abs(root)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get path")
	}
	// Try to write temp file at root
	var tempFile *os.File
	tempFile, err = ioutil.TempFile(root, "temp_")
	if err != nil {
		return nil, errors.Wrap(err, "permissions error")
	}
	defer func() {
		closeErr := tempFile.Close()
		if err == nil {
			err = closeErr
		}
		removeErr := os.Remove(tempFile.Name())
		if err == nil {
			err = removeErr
		}
	}()

	_, err = tempFile.Write([]byte("hello"))
	if err != nil {
		return nil, errors.Wrap(err, "permissions error")
	}
	return &LocalStorage{
		Root: root,
	}, nil
}

// inRoot checks whether the given path is in the root
// This is by no mean a secure check, just a convenient one (see chroot)
func (l *LocalStorage) inRoot(path string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false // Be safe and tell we are not in root path
	}
	return strings.HasPrefix(abs, l.Root)
}

// Download returns an object that can be read
func (l *LocalStorage) Download(relative string) (io.ReadCloser, error) {
	abs := filepath.Join(l.Root, filepath.FromSlash(relative))
	if !l.inRoot(abs) {
		return nil, ErrNotInRoot
	}
	return os.Open(abs)
}

// List returns a list of node in the path
func (l *LocalStorage) List(relative string) ([]StoreObject, error) {
	abs := filepath.Join(l.Root, filepath.FromSlash(relative))
	if !l.inRoot(abs) {
		return nil, errors.Wrap(ErrNotInRoot, "failed to list")
	}

	listing, err := ioutil.ReadDir(abs)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read")
	}
	nodes := make([]StoreObject, len(listing))
	for i, l := range listing {
		nodes[i] = StoreObject{
			IsDirectory: l.IsDir(),
			Modified:    l.ModTime(),
			Name:        l.Name(),
			Size:        int(l.Size()),
		}
	}
	return nodes, nil
}

// Mkdir creates a directory and potentially parents
func (l *LocalStorage) Mkdir(relative string) error {
	abs := filepath.Join(l.Root, filepath.FromSlash(relative))
	if !l.inRoot(abs) {
		return ErrNotInRoot
	}
	return os.MkdirAll(abs, defaultDirectoryPerms)
}

// Move moves a file to a new location, this can only move files and not folders
func (l *LocalStorage) Move(src, dst string) error {
	srcAbs := filepath.Join(l.Root, filepath.FromSlash(src))
	dstAbs := filepath.Join(l.Root, filepath.FromSlash(dst))
	if !l.inRoot(srcAbs) || !l.inRoot(dstAbs) {
		return ErrNotInRoot
	}

	s, err := os.Stat(srcAbs)
	if err != nil {
		return ErrNotExist
	}
	if s.IsDir() {
		return ErrDirectory
	}
	return os.Rename(srcAbs, dstAbs)
}

// Remove a path (file or empty directory)
func (l *LocalStorage) Remove(path string) error {
	abs := filepath.Join(l.Root, filepath.FromSlash(path))
	if !l.inRoot(abs) {
		return ErrNotInRoot
	}
	return os.Remove(abs)
}

// fileWithModTimeCloser implements io.WriteCloser with an underlying *os.File
// The only difference is on Close, it will call ModTime on the file too
type fileWithModTimeCloser struct {
	closed   bool
	filePath string
	file     *os.File
	modTime  time.Time
}

func (f *fileWithModTimeCloser) Write(p []byte) (n int, err error) {
	return f.file.Write(p)
}

func (f *fileWithModTimeCloser) Close() error {
	if f.closed { // Already closed
		return nil
	}
	err := f.file.Close()
	err2 := os.Chtimes(f.filePath, time.Now(), f.modTime)
	f.closed = true
	if err != nil {
		return err
	}
	return err2
}

// Upload returns an object that can be written to. It will create the
// file if it doesn't already exist. If it exists, it will overrides it
func (l *LocalStorage) Upload(path string, modTime time.Time) (io.WriteCloser, error) {
	abs := filepath.Join(l.Root, filepath.FromSlash(path))
	if !l.inRoot(abs) {
		return nil, ErrNotInRoot
	}
	f, err := os.OpenFile(abs, os.O_RDWR|os.O_CREATE, defaultPerms)
	return &fileWithModTimeCloser{
		filePath: abs,
		file:     f,
		modTime:  modTime,
	}, err
}
