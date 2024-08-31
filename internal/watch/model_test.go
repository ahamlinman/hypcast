//go:build modeltest

package watch

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed model.pml
var modelFile string

func TestWatchModel(t *testing.T) {
	var missingCmds bool
	needCmds := []string{"spin", "cc"}
	for _, cmd := range needCmds {
		if _, err := exec.LookPath(cmd); err != nil {
			t.Logf("cannot find %q on this system", cmd)
			missingCmds = true
		}
	}
	if missingCmds {
		t.SkipNow()
	}

	spinDir, err := os.MkdirTemp("", "hypcast-spin-*")
	if err != nil {
		t.Fatalf("failed to create spin compilation directory: %v", err)
	}
	t.Logf("compiling model under %v", spinDir)
	defer func() {
		if t.Failed() {
			t.Logf("keeping %v due to test failure", spinDir)
		} else {
			t.Log("cleaning up compilation directory due to successful test")
			os.RemoveAll(spinDir)
		}
	}()

	err = os.Chdir(spinDir)
	if err != nil {
		t.Fatalf("failed to change to compilation directory: %v", err)
	}

	spin := exec.Command("spin", "-a", "/dev/stdin")
	spin.Stdin = strings.NewReader(modelFile)
	spin.Stdout, spin.Stderr = os.Stdout, os.Stderr
	err = spin.Run()
	if err != nil {
		t.Fatalf("failed to run spin: %v", err)
	}

	cc := exec.Command("cc", "-o", "pan", "pan.c")
	cc.Stdout, cc.Stderr = os.Stdout, os.Stderr
	err = cc.Run()
	if err != nil {
		t.Fatalf("failed to compile pan.c: %v", err)
	}

	pan := exec.Command(filepath.Join(spinDir, "pan"))
	pan.Stdout, pan.Stderr = os.Stdout, os.Stderr
	err = pan.Run()
	if err != nil {
		t.Fatalf("failed to run pan: %v", err)
	}

	// TODO: Finish this part.
	t.Fatalf("the test does not yet know how to check for trail files")
}
