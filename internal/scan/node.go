package scan

import (
	"path/filepath"
	"sync"
	"sync/atomic"
)

// Node is a file or directory in the scanned tree.
type Node struct {
	Name    string
	Path    string
	IsDir   bool
	Parent  *Node
	Err     error
	Virtual bool // synthetic node such as "其他"

	size     atomic.Int64
	files    atomic.Int64
	dirs     atomic.Int64
	done     atomic.Bool
	scanning atomic.Bool

	mu       sync.Mutex
	children []*Node
}

func NewNode(name, path string, isDir bool, parent *Node) *Node {
	n := &Node{Name: name, Path: path, IsDir: isDir, Parent: parent}
	if isDir {
		n.scanning.Store(true)
	} else {
		n.done.Store(true)
	}
	return n
}

func NewVirtual(name, path string, parent *Node, size, files, dirs int64) *Node {
	n := NewNode(name, path, true, parent)
	n.Virtual = true
	n.size.Store(size)
	n.files.Store(files)
	n.dirs.Store(dirs)
	n.MarkDone()
	return n
}

func (n *Node) Size() int64    { return n.size.Load() }
func (n *Node) Files() int64   { return n.files.Load() }
func (n *Node) Dirs() int64    { return n.dirs.Load() }
func (n *Node) Done() bool     { return n.done.Load() }
func (n *Node) Scanning() bool { return n.scanning.Load() }

func (n *Node) AddSize(delta int64) {
	for p := n; p != nil; p = p.Parent {
		p.size.Add(delta)
	}
}

func (n *Node) AddCounts(files, dirs int64) {
	for p := n; p != nil; p = p.Parent {
		if files != 0 {
			p.files.Add(files)
		}
		if dirs != 0 {
			p.dirs.Add(dirs)
		}
	}
}

func (n *Node) AppendChild(c *Node) {
	n.mu.Lock()
	n.children = append(n.children, c)
	n.mu.Unlock()
}

func (n *Node) Children() []*Node {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]*Node, len(n.children))
	copy(out, n.children)
	return out
}

func (n *Node) ChildCount() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return len(n.children)
}

func (n *Node) MarkDone() {
	n.scanning.Store(false)
	n.done.Store(true)
}

func (n *Node) Ext() string {
	if n.IsDir {
		return ""
	}
	return filepath.Ext(n.Name)
}

func (n *Node) Depth() int {
	d := 0
	for p := n.Parent; p != nil; p = p.Parent {
		d++
	}
	return d
}
