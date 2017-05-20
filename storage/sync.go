package storage

import (
	"io"

	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

// SyncNode represents a tree node for StoreObjects
type SyncNode struct {
	StoreObject
	Children []SyncNode
}

// IsZero returns whether node is empty
func (n SyncNode) IsZero() bool {
	return n.StoreObject.IsZero() && len(n.Children) == 0
}
func (n SyncNode) String() string {
	return n.StoreObject.String()
}

// GetTree do a BFS search to generate a tree starting at root
// and generates a listing of all the nodes
func GetTree(s Storage, me StoreObject, path string) (SyncNode, error) {
	listing, err := s.List(path)
	if err != nil {
		return SyncNode{}, errors.Wrap(err, "failed to get tree")
	}

	children := make([]SyncNode, len(listing))
	for i, l := range listing {
		if l.IsDirectory {
			node, localErr := GetTree(s, l, path+"/"+l.Name)
			if err == nil && localErr != nil {
				err = localErr
			}
			children[i] = node
		} else {
			node := SyncNode{StoreObject: l}
			children[i] = node
		}
	}
	return SyncNode{
		StoreObject: me,
		Children:    children,
	}, err
}

// DiffTree returns the different tree of nodes that are in n1 but not in n2
func DiffTree(n1, n2 SyncNode) SyncNode {
	if !n1.StoreObject.Equal(n2.StoreObject) {
		return n1
	}
	children2 := make(map[string]SyncNode, len(n2.Children))
	for _, c := range n2.Children {
		children2[c.Name] = c
	}

	changedChildren := make([]SyncNode, 0, len(n1.Children))

	for _, n1Child := range n1.Children {
		if n2Child, ok := children2[n1Child.Name]; ok {
			n := DiffTree(n1Child, n2Child)
			if !n.IsZero() {
				changedChildren = append(changedChildren, n)
			}
		} else {
			changedChildren = append(changedChildren, n1Child)
		}
	}

	if len(changedChildren) == 0 {
		return SyncNode{}
	}
	return SyncNode{
		StoreObject: n1.StoreObject,
		Children:    changedChildren,
	}
}

// Sync copies everythign from src to dst. If there are more things in dst,
// move them to the Bin
// Bin -> ToDo
func Sync(src Storage, srcRoot string, dst Storage, dstRoot string) error {
	srcRootObj := StoreObject{}
	// ToDo: Cleaner error check
	if _, err := src.List(srcRoot); err == nil {
		srcRootObj.IsDirectory = true
	}
	dstRootObj := StoreObject{}
	if _, err := dst.List(dstRoot); err == nil {
		dstRootObj.IsDirectory = true
	}

	srcTree, err := GetTree(src, srcRootObj, srcRoot)
	if err != nil {
		return err
	}
	dstTree, err2 := GetTree(dst, dstRootObj, dstRoot)
	if err2 != nil {
		return err2
	}
	diff := DiffTree(srcTree, dstTree)
	if diff.IsZero() { // Nothing to do
		log.Info("Directories are in sync")
		return nil
	}
	var dfsWalk func(n SyncNode, srcPath, dstPath string) error
	dfsWalk = func(n SyncNode, srcPath, dstPath string) error {
		srcPath = srcPath + "/" + n.Name
		dstPath = dstPath + "/" + n.Name
		if !n.IsDirectory {
			log.Infof("Copying %s", dstPath)
			srcFile, err := src.Download(srcPath)
			if err != nil {
				return errors.Wrap(err, "failed to open "+srcPath)
			}
			defer srcFile.Close()
			dstFile, err := dst.Upload(dstPath, n.Modified)
			if err != nil {
				return errors.Wrap(err, "failed to open "+dstPath)
			}
			defer dstFile.Close()
			_, err = io.Copy(dstFile, srcFile)
			if err != nil {
				return errors.Wrap(err, "failed to copy "+srcPath+" to "+dstPath)
			}
		} else {
			err := dst.Mkdir(dstPath)
			if err != nil {
				return errors.Wrap(err, "failed to create directory "+dstPath)
			}
			for _, c := range n.Children {
				err = dfsWalk(c, srcPath, dstPath)
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	return dfsWalk(diff, srcRoot, dstRoot)
}
