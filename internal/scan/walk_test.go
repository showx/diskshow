package scan

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestScanSelf(t *testing.T) {
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	root = filepath.Dir(root) // internal/
	root = filepath.Dir(root) // repo root

	sc := NewScanner(root, Options{SkipSpecial: true})
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	sc.Start(ctx)
	sc.Wait()

	st := sc.Snapshot()
	if !st.Done {
		t.Fatal("scan should be done")
	}
	if st.Files == 0 {
		t.Fatal("expected some files")
	}
	if sc.Root().Size() <= 0 {
		t.Fatal("expected nonzero size")
	}
}
