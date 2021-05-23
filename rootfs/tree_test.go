package rootfs

import (
	"testing"
)

func TestTree(t *testing.T) {
	tests := []struct {
		name       string
		paths      []string
		matchTrue  []string
		matchFalse []string
	}{
		{
			name:       "empty sequence matches nothing",
			paths:      []string{},
			matchFalse: []string{"a", "b"},
		},
		{
			name:      "directory",
			paths:     []string{"a/b"},
			matchTrue: []string{"a/b/"},
		},
		{
			name:       "a few sequences",
			paths:      []string{"a", "b", "c/b/a"},
			matchTrue:  []string{"a", "a/b/c", "c/b/a", "c/b/a/d"},
			matchFalse: []string{"c/d", "c", "c/b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := newTree(tt.paths...)

			for _, path := range tt.matchTrue {
				t.Run(path, func(t *testing.T) {
					if !tree.HasPrefix(path) {
						t.Errorf("expected %s to be a prefix of %s", path, tree)
					}
				})
			}

			for _, path := range tt.matchFalse {
				t.Run(path, func(t *testing.T) {
					if tree.HasPrefix(path) {
						t.Errorf("expected %s to not be a prefix of %s", path, tree)
					}
				})
			}
		})
	}
}

func TestTreeMerge(t *testing.T) {
	tree1 := newTree("bin/ar", "var/cache/apt")
	tree2 := newTree("bin/ar", "bin/busybox", "usr/share/doc")
	tree1.Merge(tree2)
	want1 := "./bin/ar:./bin/busybox:./usr/share/doc:./var/cache/apt"
	if got := tree1.String(); want1 != got {
		t.Errorf("want != got: %q != %q", want1, got)
	}
	want2 := "./bin/ar:./bin/busybox:./usr/share/doc"
	if got := tree2.String(); want2 != got {
		t.Errorf("want != got: %q != %q", want2, got)
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		paths   []string
		wantStr string
	}{
		{
			name:    "empty",
			paths:   []string{},
			wantStr: "<empty>",
		},
		{
			name:    "simple path",
			paths:   []string{"a/b/c"},
			wantStr: "./a/b/c",
		},
		{
			name:    "duplicate paths",
			paths:   []string{"a/a", "a//a"},
			wantStr: "./a/a",
		},
		{
			name:    "a few sequences",
			paths:   []string{"bin/ar", "bin/busybox", "var/cache/apt/archives"},
			wantStr: "./bin/ar:./bin/busybox:./var/cache/apt/archives",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tree := newTree(tt.paths...)
			if tt.wantStr != tree.String() {
				t.Errorf("want != got: %q != %q", tt.wantStr, tree.String())
			}
		})
	}
}
