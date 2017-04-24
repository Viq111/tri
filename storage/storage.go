package storage

import (
	"io"
	"time"
)

// StoreObject defines an object in the storage. Name is relative to current path
type StoreObject struct {
	Directory bool
	Modified  time.Time
	Name      string
	Size      int
}

// IsZero returns whether StoreObject is an empty object
func (s StoreObject) IsZero() bool {
	return s == StoreObject{}
}

func (s StoreObject) String() string {
	return s.Name
}

type sortAlphabetical []StoreObject

func (n sortAlphabetical) Len() int           { return len(n) }
func (n sortAlphabetical) Less(i, j int) bool { return n[i].Name < n[j].Name }
func (n sortAlphabetical) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

// Storage defines the base method that any storage should implement
// It only defines an interface to backup data
type Storage interface {
	Download(path string) (io.ReadCloser, error)
	List(path string) ([]StoreObject, error)
	Mkdir(path string) error
	Move(src, dst string) error
	Remove(path string) error
	Upload(path string) (io.WriteCloser, error)
}
