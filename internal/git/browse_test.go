package git

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestBrowseRecords(t *testing.T) {
	ctx := context.Background()
	repo := Workspace{Repository: Repository{Root: t.TempDir()}}
	repo.Runner = &fakeRunner{results: [][]byte{[]byte("abc\x002026-11-22T10:00:00Z\x00A Ω\x00two\twords\x00")}}
	page, err := repo.Log(ctx, Branch{OID: "abc"}, "-file", 1, 2)
	if err != nil || len(page) != 1 || page[0].Subject != "two\twords" {
		t.Fatalf("%+v %v", page, err)
	}
	repo.Runner = &fakeRunner{results: [][]byte{[]byte("100644 aa 1\tline\nbreak\x00100644 bb 3\tline\nbreak\x00")}}
	conflicts, err := repo.Conflicts(ctx)
	if err != nil || len(conflicts) != 1 || conflicts[0].Path != "line\nbreak" || conflicts[0].Stages[1] != "aa" || conflicts[0].Stages[2] != "" || conflicts[0].Stages[3] != "bb" {
		t.Fatalf("%+v %v", conflicts, err)
	}
	repo.Runner = &fakeRunner{err: ErrOutputLimit}
	_, large, err := repo.Patch(ctx, "", "", false)
	if err != nil || !large {
		t.Fatalf("large=%t %v", large, err)
	}
	repo.Runner = &fakeRunner{err: context.Canceled}
	_, _, err = repo.Patch(ctx, "", "", false)
	if !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
func TestNestedRepositoryBoundary(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "inner", ".git"), 0700); err != nil {
		t.Fatal(err)
	}
	repo := Workspace{Repository: Repository{Root: dir}}
	if _, err := repo.PathInfo("inner/file", true); err == nil {
		t.Fatal("nested checkout accepted")
	}
}
