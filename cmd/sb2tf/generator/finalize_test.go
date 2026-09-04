package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFileWithResourceCount writes a minimal .tf file to dir/name containing
// count distinct `resource "..." "..." {}` blocks, for Finalize's shrink
// guard to count against.
func writeFileWithResourceCount(t *testing.T, dir, name string, count int) {
	t.Helper()
	var buf bytes.Buffer
	for i := 0; i < count; i++ {
		buf.WriteString("resource \"stonebranch_task_unix\" \"task_unix_00")
		buf.WriteString(string(rune('1' + i)))
		buf.WriteString("\" {\n  name = \"x\"\n}\n\n")
	}
	if err := os.WriteFile(filepath.Join(dir, name), buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to seed %s: %v", name, err)
	}
}

func TestFinalize_RefusesShrinkingOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	writeFileWithResourceCount(t, dir, "tasks_unix.tf", 3)

	g := NewGenerator(nil, dir, false, false, false)
	buf := &bytes.Buffer{}
	buf.WriteString("resource \"stonebranch_task_unix\" \"task_unix_001\" {\n  name = \"x\"\n}\n\n")
	g.buffers["tasks_unix.tf"] = buf

	err := g.Finalize()
	if err == nil {
		t.Fatal("expected Finalize to refuse the shrinking overwrite, got nil error")
	}
	if !strings.Contains(err.Error(), "tasks_unix.tf") {
		t.Fatalf("expected error to mention tasks_unix.tf, got: %v", err)
	}

	// The on-disk file must be untouched (still 3 resources), not truncated
	// to the 1 resource this run produced.
	content, readErr := os.ReadFile(filepath.Join(dir, "tasks_unix.tf"))
	if readErr != nil {
		t.Fatalf("failed to read back tasks_unix.tf: %v", readErr)
	}
	if got := countResourceBlocks(string(content)); got != 3 {
		t.Fatalf("expected on-disk file to remain untouched with 3 resources, got %d", got)
	}
}

func TestFinalize_ForceOverridesShrinkGuard(t *testing.T) {
	dir := t.TempDir()
	writeFileWithResourceCount(t, dir, "tasks_unix.tf", 3)

	g := NewGenerator(nil, dir, false, false, true)
	buf := &bytes.Buffer{}
	buf.WriteString("resource \"stonebranch_task_unix\" \"task_unix_001\" {\n  name = \"x\"\n}\n\n")
	g.buffers["tasks_unix.tf"] = buf

	if err := g.Finalize(); err != nil {
		t.Fatalf("expected --force to allow the overwrite, got error: %v", err)
	}

	content, readErr := os.ReadFile(filepath.Join(dir, "tasks_unix.tf"))
	if readErr != nil {
		t.Fatalf("failed to read back tasks_unix.tf: %v", readErr)
	}
	if got := countResourceBlocks(string(content)); got != 1 {
		t.Fatalf("expected forced overwrite to have 1 resource, got %d", got)
	}
}

func TestFinalize_AllowsGrowingOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	writeFileWithResourceCount(t, dir, "tasks_unix.tf", 1)

	g := NewGenerator(nil, dir, false, false, false)
	buf := &bytes.Buffer{}
	buf.WriteString("resource \"stonebranch_task_unix\" \"task_unix_001\" {\n  name = \"x\"\n}\n\n")
	buf.WriteString("resource \"stonebranch_task_unix\" \"task_unix_002\" {\n  name = \"y\"\n}\n\n")
	g.buffers["tasks_unix.tf"] = buf

	if err := g.Finalize(); err != nil {
		t.Fatalf("expected a growing overwrite to succeed without --force, got error: %v", err)
	}

	content, readErr := os.ReadFile(filepath.Join(dir, "tasks_unix.tf"))
	if readErr != nil {
		t.Fatalf("failed to read back tasks_unix.tf: %v", readErr)
	}
	if got := countResourceBlocks(string(content)); got != 2 {
		t.Fatalf("expected 2 resources after growing overwrite, got %d", got)
	}
}
