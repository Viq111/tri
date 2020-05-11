package tests

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

/*
Usage:

tf := NewLocalTestFolder(t)
defer tf.Cleanup()
in, out := tf.GetInputOutputDir()
// Populate in directory
// Run whatever command / operation you want
tf.DirectoriesEqual() // Check if directories are equal, fail testing if not
*/

type LocalTest struct {
	inputDirectory, outputDirectory string
	root                            string
	t                               *testing.T
}

// NewLocalTestFolder creates a new directory
func NewLocalTestFolder(t *testing.T) *LocalTest {
	root, err := ioutil.TempDir("", "tri_integration_tests_*")
	if err != nil {
		t.Fatalf("Failed to create local test folder: %s", err)
	}
	// Create 2 subdirs
	in := filepath.Join(root, "input")
	err = os.Mkdir(in, 0660)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %s", err)
	}
	out := filepath.Join(root, "output")
	err = os.Mkdir(out, 0660)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %s", err)
	}

	return &LocalTest{
		inputDirectory:  in,
		outputDirectory: out,
		root:            root,
		t:               t,
	}
}

// Cleanup destroys the local test folder at the end of testing
func (lt *LocalTest) Cleanup() {
	err := os.RemoveAll(lt.root)
	if err != nil {
		// Only log remove failure but don't fail
		lt.t.Logf("Failed to cleanup test directory: %s", err)
	}
}

// GetInputOutputDir returns the directories you can use to write input and output
func (lt *LocalTest) GetInputOutputDir() (string, string) {
	return lt.inputDirectory, lt.outputDirectory
}

// DirectoriesEqual checks if directories are equal. It calls t.Error if not
func (lt *LocalTest) DirectoriesEqual() {
	cmd := exec.Command("diff", "-r", lt.inputDirectory, lt.outputDirectory)
	out, err := cmd.CombinedOutput()
	if err != nil || len(out) > 0 {
		lt.t.Errorf("failed comparing directories (err=%s), output:\n%s", err, out)
	}
}
