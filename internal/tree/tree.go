package tree

import (
	"path/filepath"
	"sort"
	"strings"
)

// Tree is a way to store directory paths for whiteouts.
// It is semi-optimized for reads and non-optimized for writes;
// See Merge() and HasPrefix for trade-offs.
type Tree struct {
	name     string
	children []*Tree
	end      bool
}

// New creates a new tree from a given path.
func New(paths ...string) *Tree {
	t := &Tree{name: ".", children: []*Tree{}}
	for _, path := range paths {
		t.Add(path)
	}
	return t
}

// Add adds a sequence to a tree
func (t *Tree) Add(path string) {
	t.add(strings.Split(filepath.Clean(path), "/"))
}

// HasPrefix returns if tree contains a prefix matching a given sequence.
// Search algorithm is naive: it does linear search when going through the
// nodes instead of binary-search. Since we expect number of children to be
// really small (usually 1 or 2), it does not really matter. If you find a
// real-world container with 30+ whiteout paths on a single path, please ping
// the author/maintainer of this code.
func (t *Tree) HasPrefix(path string) bool {
	return t.hasprefix(strings.Split(filepath.Clean(path), "/"))
}

// Merge merges adds t2 to t. It is not optimized for speed, since it's walking
// full branch for every other branch.
func (t *Tree) Merge(t2 *Tree) {
	t.merge(t2, []string{})
}

// String stringifies a tree
func (t *Tree) String() string {
	if len(t.children) == 0 {
		return "<empty>"
	}

	res := &stringer{[]string{}}
	res.stringify(t, []string{})
	sort.Strings(res.res)
	return strings.Join(res.res, ":")
}

func (t *Tree) add(nodes []string) {
	if len(nodes) == 0 {
		t.end = true
		return
	}
	for i := range t.children {
		if t.children[i].name == nodes[0] {
			t.children[i].add(nodes[1:])
			return
		}
	}

	newNode := &Tree{name: nodes[0]}
	t.children = append(t.children, newNode)
	newNode.add(nodes[1:])
}

func (t *Tree) hasprefix(nodes []string) bool {
	if len(nodes) == 0 {
		return t.end
	}
	if t.end {
		return true
	}

	for i := range t.children {
		if t.children[i].name == nodes[0] {
			return t.children[i].hasprefix(nodes[1:])
		}
	}

	return false
}

type stringer struct {
	res []string
}

func (s *stringer) stringify(t *Tree, acc []string) {
	if t.name == "" {
		return
	}
	acc = append(acc, t.name)
	if t.end {
		s.res = append(s.res, strings.Join(acc, "/"))
	}

	for _, child := range t.children {
		s.stringify(child, acc)
	}
}

func (t *Tree) merge(t2 *Tree, acc []string) {
	if t2.end {
		t.add(append(acc[1:], t2.name))
	}
	acc = append(acc, t2.name)
	for _, child := range t2.children {
		t.merge(child, acc)
	}
}
