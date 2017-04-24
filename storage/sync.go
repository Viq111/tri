package storage

import (
	"fmt"
	"io"
	"sort"

	"github.com/pkg/errors"
)

// Node represents a tree node for StoreObjects
type Node struct {
	StoreObject
	Children []Node
}

// IsZero returns whether node is empty
func (n Node) IsZero() bool {
	return n.StoreObject.IsZero() && len(n.Children) == 0
}

type sortNodeAlphabetical []Node

func (n sortNodeAlphabetical) Len() int           { return len(n) }
func (n sortNodeAlphabetical) Less(i, j int) bool { return n[i].Name < n[j].Name }
func (n sortNodeAlphabetical) Swap(i, j int)      { n[i], n[j] = n[j], n[i] }

// GetTree do a BFS search to generate a tree starting at root
// and generates a listing of all the nodes
func GetTree(s Storage, me StoreObject, path string) (Node, error) {
	listing, err := s.List(path)
	if err != nil {
		return Node{}, errors.Wrap(err, "failed to get tree")
	}

	children := make([]Node, len(listing))
	for i, l := range listing {
		if l.Directory {
			node, localErr := GetTree(s, l, path+"/"+l.Name)
			if err == nil && localErr != nil {
				err = localErr
			}
			children[i] = node
		} else {
			node := Node{StoreObject: l}
			children[i] = node
		}
	}
	return Node{
		StoreObject: me,
		Children:    children,
	}, err
}

// DiffTree returns the different tree of nodes that are in n1 but not in n2
func DiffTree(n1, n2 Node) Node {
	if n1.StoreObject != n2.StoreObject {
		return n1
	}
	sort.Sort(sortNodeAlphabetical(n1.Children))
	sort.Sort(sortNodeAlphabetical(n2.Children))
	changedChildren := make([]Node, 0, len(n1.Children))

	for i, n1Child := range n1.Children {
		if i >= len(n2.Children) {
			changedChildren = append(changedChildren, n1Child)
			continue
		}
		n := DiffTree(n1Child, n2.Children[i])
		if !n.IsZero() {
			changedChildren = append(changedChildren, n)
		}
	}

	if len(changedChildren) == 0 {
		return Node{}
	}
	return Node{
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
		srcRootObj.Directory = true
	}
	dstRootObj := StoreObject{}
	if _, err := dst.List(dstRoot); err == nil {
		dstRootObj.Directory = true
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
		fmt.Println("Directories are in sync")
	}
	var dfsWalk func(n Node, srcPath, dstPath string) error
	dfsWalk = func(n Node, srcPath, dstPath string) error {
		srcPath = srcPath + "/" + n.Name
		dstPath = dstPath + "/" + n.Name
		if !n.Directory {
			srcFile, err := src.Download(srcPath)
			if err != nil {
				return errors.Wrap(err, "failed to open "+srcPath)
			}
			defer srcFile.Close()
			dstFile, err := dst.Upload(dstPath)
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
