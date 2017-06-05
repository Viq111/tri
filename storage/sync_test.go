package storage

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSyncGetTree(t *testing.T) {
	assert := assert.New(t)
	root := StoreObject{}

	tree, err := GetTree(storageTestGlobal.GetStorage(), root, ".")
	assert.NoError(err, "failed to get tree")
	assert.Len(tree.Children, 3)
	sort.Sort(sortAlphabeticalSyncNode(tree.Children))
	assert.Equal("file_a", tree.Children[0].Name)
	assert.Equal(false, tree.Children[0].IsDirectory)
	assert.Equal("folder_a", tree.Children[1].Name)
	assert.Equal(true, tree.Children[1].IsDirectory)
	assert.Equal("folder_empty", tree.Children[2].Name)
	assert.Equal(true, tree.Children[2].IsDirectory)

	folder_a := tree.Children[1]
	sort.Sort(sortAlphabeticalSyncNode(folder_a.Children))
	assert.Equal("folder_b", folder_a.Children[0].Name)
	assert.Equal(true, folder_a.Children[0].IsDirectory)
	assert.Equal("folder_empty", folder_a.Children[1].Name)
	assert.Equal(true, folder_a.Children[1].IsDirectory)

	assert.Len(folder_a.Children[0].Children, 1)
	assert.Equal("file_b", folder_a.Children[0].Children[0].Name)
	assert.Equal(false, folder_a.Children[0].Children[0].IsDirectory)
	assert.Len(folder_a.Children[1].Children, 0)
}

func TestSyncDiffTree(t *testing.T) {
	assert := assert.New(t)
	/*
				/file_a
				/file_b
				/file_c
				/folder_a/file_aa
				/folder_a/file_ab
				/folder_b/file_ba

				VS

				/file_a (not same date)
				/file_c (not same size)
				/file_d
				/folder_a/file_aa
				/folder_b/file_ba

				should return
				/file_a
				/file_b
		        /file_c
				/folder_a/file_ab
	*/
	date1 := time.Date(2017, time.January, 10, 9, 55, 3, 0, time.UTC)
	date2 := time.Date(2017, time.January, 10, 8, 55, 3, 0, time.UTC)

	node1 := SyncNode{
		StoreObject: StoreObject{Name: "/", IsDirectory: true},
		Children: []SyncNode{
			SyncNode{
				StoreObject: StoreObject{Name: "file_a", Size: 10, Modified: date1},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "file_b", Size: 10, Modified: date1},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "file_c", Size: 10, Modified: date1},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "folder_a", IsDirectory: true},
				Children: []SyncNode{
					SyncNode{
						StoreObject: StoreObject{Name: "file_aa", Size: 10, Modified: date1},
					},
					SyncNode{
						StoreObject: StoreObject{Name: "file_ab", Size: 10, Modified: date1},
					},
				},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "folder_b", IsDirectory: true},
				Children: []SyncNode{
					SyncNode{
						StoreObject: StoreObject{Name: "file_ba", Size: 10, Modified: date1},
					},
				},
			},
		},
	}

	node2 := SyncNode{
		StoreObject: StoreObject{Name: "/", IsDirectory: true},
		Children: []SyncNode{
			SyncNode{
				StoreObject: StoreObject{Name: "file_a", Size: 10, Modified: date2},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "file_b", Size: 25, Modified: date1},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "file_d", Size: 10, Modified: date1},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "folder_a", IsDirectory: true},
				Children: []SyncNode{
					SyncNode{
						StoreObject: StoreObject{Name: "file_aa", Size: 10, Modified: date1},
					},
				},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "folder_b", IsDirectory: true},
				Children: []SyncNode{
					SyncNode{
						StoreObject: StoreObject{Name: "file_ba", Size: 10, Modified: date1},
					},
				},
			},
		},
	}

	expected := SyncNode{
		StoreObject: StoreObject{Name: "/", IsDirectory: true},
		Children: []SyncNode{
			SyncNode{
				StoreObject: StoreObject{Name: "file_a", Size: 10, Modified: date1},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "file_b", Size: 10, Modified: date1},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "file_c", Size: 10, Modified: date1},
			},
			SyncNode{
				StoreObject: StoreObject{Name: "folder_a", IsDirectory: true},
				Children: []SyncNode{
					SyncNode{
						StoreObject: StoreObject{Name: "file_ab", Size: 10, Modified: date1},
					},
				},
			},
		},
	}

	diff := DiffTree(node1, node2)
	assert.Equal(expected, diff)
}
