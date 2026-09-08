package scan

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	Files      int64
	Dirs       int64
	Bytes      int64
	Errors     int64
	Current    string
	StartedAt  time.Time
	FinishedAt time.Time
	Done       bool
}

type Scanner struct {
	root        *Node
	rootPath    string
	skipHidden  bool
	skipSpecial bool

	mu      sync.RWMutex
	current string
	errN    atomic.Int64
	done    atomic.Bool
	started time.Time
	ended   time.Time

	sem chan struct{}
	wg  sync.WaitGroup
}

type Options struct {
	SkipHidden  bool
	SkipSpecial bool
	Workers     int
}

func NewScanner(rootPath string, opt Options) *Scanner {
	abs, err := filepath.Abs(rootPath)
	if err != nil {
		abs = rootPath
	}
	workers := opt.Workers
	if workers <= 0 {
		workers = runtime.NumCPU() * 4
		if workers < 8 {
			workers = 8
		}
		if workers > 64 {
			workers = 64
		}
	}
	name := filepath.Base(abs)
	if name == "." || name == string(filepath.Separator) || name == "" {
		name = abs
	}
	root := NewNode(name, abs, true, nil)
	return &Scanner{
		root:        root,
		rootPath:    abs,
		skipHidden:  opt.SkipHidden,
		skipSpecial: opt.SkipSpecial,
		sem:         make(chan struct{}, workers),
	}
}

func (s *Scanner) Root() *Node { return s.root }

func (s *Scanner) Start(ctx context.Context) {
	s.started = time.Now()
	s.setCurrent(s.rootPath)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.walkDir(ctx, s.root)
		s.ended = time.Now()
		s.done.Store(true)
		s.setCurrent("")
	}()
}

func (s *Scanner) Wait() { s.wg.Wait() }

func (s *Scanner) Snapshot() Stats {
	s.mu.RLock()
	cur := s.current
	s.mu.RUnlock()
	st := Stats{
		Files:     s.root.Files(),
		Dirs:      s.root.Dirs(),
		Bytes:     s.root.Size(),
		Errors:    s.errN.Load(),
		Current:   cur,
		StartedAt: s.started,
		Done:      s.done.Load(),
	}
	if st.Done {
		st.FinishedAt = s.ended
	}
	return st
}

func (s *Scanner) setCurrent(p string) {
	s.mu.Lock()
	s.current = p
	s.mu.Unlock()
}

func (s *Scanner) walkDir(ctx context.Context, node *Node) {
	defer node.MarkDone()
	if ctx.Err() != nil {
		return
	}
	s.setCurrent(node.Path)

	s.sem <- struct{}{}
	entries, err := os.ReadDir(openPath(node.Path))
	<-s.sem
	if err != nil {
		node.Err = err
		s.errN.Add(1)
		return
	}

	var dirs []*Node
	for _, de := range entries {
		if ctx.Err() != nil {
			break
		}
		name := de.Name()
		childPath := join(node.Path, name)
		if s.skipHidden && isHiddenName(name) {
			continue
		}
		if shouldSkip(childPath, de, s.skipSpecial) {
			continue
		}

		if de.IsDir() {
			child := NewNode(name, childPath, true, node)
			node.AppendChild(child)
			node.AddCounts(0, 1)
			dirs = append(dirs, child)
			continue
		}

		info, err := de.Info()
		if err != nil {
			s.errN.Add(1)
			continue
		}
		if info.Mode()&os.ModeSymlink != 0 {
			continue
		}
		sz := info.Size()
		child := NewNode(name, childPath, false, node)
		child.size.Store(sz)
		node.AppendChild(child)
		node.AddSize(sz)
		node.AddCounts(1, 0)
	}

	var dirWg sync.WaitGroup
	for _, child := range dirs {
		if ctx.Err() != nil {
			break
		}
		dirWg.Add(1)
		go func(c *Node) {
			defer dirWg.Done()
			s.walkDir(ctx, c)
		}(child)
	}
	dirWg.Wait()
}

func openPath(p string) string {
	if runtime.GOOS != "windows" {
		return p
	}
	if len(p) < 248 {
		return p
	}
	if strings.HasPrefix(p, `\\?\`) {
		return p
	}
	if strings.HasPrefix(p, `\\`) {
		return `\\?\UNC\` + strings.TrimPrefix(p, `\\`)
	}
	return `\\?\` + p
}
