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
	for _, cmd := range []string{"spin", "cc"} {
		if _, err := exec.LookPath(cmd); err != nil {
			t.Fatalf("cannot find %v on this system", cmd)
		}
	}

	tmpdir, err := os.MkdirTemp("", "hypcast-spin-*")
	if err != nil {
		t.Fatalf("failed to create spin compilation directory: %v", err)
	}

	t.Logf("compiling model under %v", tmpdir)
	defer func() {
		if t.Failed() {
			t.Logf("keeping %v due to test failure", tmpdir)
			return
		}
		if err := os.RemoveAll(tmpdir); err == nil {
			t.Logf("cleaned up %v", tmpdir)
		} else {
			t.Logf("failed to clean up %v", tmpdir)
		}
	}()

	if err := os.Chdir(tmpdir); err != nil {
		t.Fatalf("failed to change to compilation directory: %v", err)
	}

	spin := exec.Command("spin", "-a", "/dev/stdin")
	spin.Stdin = strings.NewReader(modelFile)
	spin.Stdout, spin.Stderr = os.Stdout, os.Stderr
	if err := spin.Run(); err != nil {
		t.Fatalf("failed to run spin: %v", err)
	}

	cc := exec.Command("cc", "-o", "pan", "pan.c")
	cc.Stdout, cc.Stderr = os.Stdout, os.Stderr
	if err := cc.Run(); err != nil {
		t.Fatalf("failed to compile pan.c: %v", err)
	}

	pan := exec.Command(filepath.Join(tmpdir, "pan"))
	pan.Stdout, pan.Stderr = os.Stdout, os.Stderr
	if err := pan.Run(); err != nil {
		t.Fatalf("failed to run pan: %v", err)
	}

	matches, _ := filepath.Glob("*.trail") // Error-free for well-formed patterns.
	if len(matches) == 0 {
		return
	}

	t.Fatalf("found trail files: %v", matches) // TODO: Print the trail.
}
