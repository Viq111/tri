package tests

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func fatalOnError(t *testing.T, err error, msg []byte) {
	if err != nil {
		t.Fatalf("Failed: err=%s, msg=%s", err, msg)
	}
}

func TestMain(m *testing.M) {
	// Set up tri in cli
	_, err := exec.Command("go", "install", "github.com/Viq111/tri").CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to run go install: %s\n", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestSimple(t *testing.T) {
	t.Run("Simple backup", func(t *testing.T) {
		tf := NewLocalTestFolder(t)
		defer tf.Cleanup()
		in, out := tf.GetInputOutputDir()
		stdout, err := exec.Command("git", "clone", "https://github.com/Viq111/iqredisio.git", in).CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to git clone: err=%s, msg=%s", err, stdout)
		}
		stdout, err = exec.Command("tri", "sync", in, out).CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to invoke tri: err=%s, msg=%s", err, stdout)
		}
		// Compare
		tf.DirectoriesEqual()
	})
}
