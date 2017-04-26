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

// Equal test the equality of 2 store objects based
// on only available (non-zero) fields
func (s StoreObject) Equal(other StoreObject) bool {
	if s.Directory != other.Directory {
		return false
	}
	if !s.Modified.IsZero() && !other.Modified.IsZero() && s.Modified != other.Modified {
		//return false
	}
	if s.Size != 0 && other.Size != 0 && s.Size != other.Size {
		return false
	}
	return s.Name == other.Name
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
