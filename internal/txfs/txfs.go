package txfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type changeKind int

const (
	changeReplace changeKind = iota + 1
	changeDelete
)

// FS is a transactional overlay filesystem rooted at workspace root.
// Reads prefer overlay data; writes are captured in overlay until Commit.
type FS struct {
	root        string
	overlay     string
	changes     map[string]changeKind
	tombstone   map[string]struct{}
	changeIndex pathTrie
	deleteIndex pathTrie
	overlayDirs map[string]struct{}
}

func New(root, overlay string) (*FS, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	absOverlay, err := filepath.Abs(overlay)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absOverlay, 0o755); err != nil {
		return nil, err
	}
	return &FS{
		root:        absRoot,
		overlay:     absOverlay,
		changes:     map[string]changeKind{},
		tombstone:   map[string]struct{}{},
		overlayDirs: map[string]struct{}{absOverlay: {}},
	}, nil
}

func (fs *FS) Open(name string) (*os.File, error) {
	fs.ensureIndexes()
	rel, err := cleanRel(name)
	if err != nil {
		return nil, err
	}
	if fs.isDeleted(rel) {
		return nil, os.ErrNotExist
	}
	if fs.changes[rel] == changeReplace {
		overlayPath := filepath.Join(fs.overlay, rel)
		if f, err := os.Open(overlayPath); err == nil {
			return f, nil
		}
	}
	return os.Open(filepath.Join(fs.root, rel))
}

func (fs *FS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	fs.ensureIndexes()
	rel, err := cleanRel(name)
	if err != nil {
		return nil, err
	}
	overlayPath := filepath.Join(fs.overlay, rel)
	if err := fs.ensureOverlayDir(filepath.Dir(overlayPath)); err != nil {
		return nil, err
	}
	if err := fs.materializeForWrite(rel, flag, perm); err != nil {
		return nil, err
	}
	delete(fs.tombstone, rel)
	fs.deleteIndex.Remove(rel)
	fs.setChange(rel, changeReplace)
	return os.OpenFile(overlayPath, flag, perm)
}

func (fs *FS) MkdirAll(path string, perm os.FileMode) error {
	fs.ensureIndexes()
	rel, err := cleanRel(path)
	if err != nil {
		return err
	}
	delete(fs.tombstone, rel)
	fs.deleteIndex.Remove(rel)
	if err := os.MkdirAll(filepath.Join(fs.overlay, rel), perm); err != nil {
		return err
	}
	fs.setChange(rel, changeReplace)
	fs.overlayDirs[filepath.Clean(filepath.Join(fs.overlay, rel))] = struct{}{}
	return nil
}

func (fs *FS) Remove(name string) error {
	fs.ensureIndexes()
	rel, err := cleanRel(name)
	if err != nil {
		return err
	}
	fs.markDelete(rel)
	_ = os.Remove(filepath.Join(fs.overlay, rel))
	return nil
}

func (fs *FS) RemoveAll(path string) error {
	fs.ensureIndexes()
	rel, err := cleanRel(path)
	if err != nil {
		return err
	}
	fs.markDelete(rel)
	_ = os.RemoveAll(filepath.Join(fs.overlay, rel))
	return nil
}

func (fs *FS) Rename(oldPath, newPath string) error {
	fs.ensureIndexes()
	oldRel, err := cleanRel(oldPath)
	if err != nil {
		return err
	}
	newRel, err := cleanRel(newPath)
	if err != nil {
		return err
	}
	if err := fs.materializeForWrite(oldRel, os.O_RDWR, 0o644); err != nil {
		return err
	}
	oldOverlay := filepath.Join(fs.overlay, oldRel)
	newOverlay := filepath.Join(fs.overlay, newRel)
	if err := os.MkdirAll(filepath.Dir(newOverlay), 0o755); err != nil {
		return err
	}
	if err := os.Rename(oldOverlay, newOverlay); err != nil {
		return err
	}
	fs.markDelete(oldRel)
	delete(fs.tombstone, newRel)
	fs.deleteIndex.Remove(newRel)
	fs.setChange(newRel, changeReplace)
	return nil
}

func (fs *FS) Stat(name string) (os.FileInfo, error) {
	fs.ensureIndexes()
	rel, err := cleanRel(name)
	if err != nil {
		return nil, err
	}
	if fs.isDeleted(rel) {
		return nil, os.ErrNotExist
	}
	if fs.changes[rel] == changeReplace {
		overlayPath := filepath.Join(fs.overlay, rel)
		if info, err := os.Stat(overlayPath); err == nil {
			return info, nil
		}
	}
	return os.Stat(filepath.Join(fs.root, rel))
}

func (fs *FS) Commit() error {
	fs.ensureIndexes()
	for rel, kind := range fs.changes {
		target := filepath.Join(fs.root, rel)
		switch kind {
		case changeDelete:
			if err := os.RemoveAll(target); err != nil && !os.IsNotExist(err) {
				return err
			}
		case changeReplace:
			source := filepath.Join(fs.overlay, rel)
			info, err := os.Stat(source)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			if info.IsDir() {
				if err := os.MkdirAll(target, info.Mode().Perm()); err != nil {
					return err
				}
				continue
			}
			if err := replaceFileStreaming(source, target, info.Mode().Perm()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (fs *FS) Rollback() error {
	return os.RemoveAll(fs.overlay)
}

func (fs *FS) materializeForWrite(rel string, flag int, perm os.FileMode) error {
	fs.ensureIndexes()
	overlayPath := filepath.Join(fs.overlay, rel)
	if fs.changes[rel] == changeReplace {
		return nil
	}
	if flag&os.O_TRUNC != 0 || flag&os.O_CREATE != 0 {
		return nil
	}
	basePath := filepath.Join(fs.root, rel)
	if !exists(basePath) {
		return nil
	}
	if err := copyFileStreaming(basePath, overlayPath, perm); err != nil {
		return err
	}
	fs.setChange(rel, changeReplace)
	return nil
}

func (fs *FS) isDeleted(rel string) bool {
	fs.ensureIndexes()
	return fs.deleteIndex.HasPrefix(rel)
}

func (fs *FS) markDelete(rel string) {
	fs.ensureIndexes()
	deletedDescendants := fs.deleteIndex.Descendants(rel)
	fs.deleteIndex.MarkDeleted(rel)
	fs.tombstone[rel] = struct{}{}
	for _, p := range fs.changeIndex.Descendants(rel) {
		if p == rel {
			continue
		}
		delete(fs.changes, p)
		fs.changeIndex.Remove(p)
	}
	for _, p := range deletedDescendants {
		if p == rel {
			continue
		}
		delete(fs.tombstone, p)
	}
	fs.setChange(rel, changeDelete)
}

func cleanRel(name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty path")
	}
	clean := filepath.Clean(name)
	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("absolute path is not allowed: %q", name)
	}
	if clean == "." || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path escapes workspace root: %q", name)
	}
	return clean, nil
}

func (fs *FS) ensureIndexes() {
	if fs.changes == nil {
		fs.changes = map[string]changeKind{}
	}
	if fs.tombstone == nil {
		fs.tombstone = map[string]struct{}{}
	}
	if fs.overlayDirs == nil {
		fs.overlayDirs = map[string]struct{}{fs.overlay: {}}
	}
	if fs.changeIndex.Empty() && len(fs.changes) > 0 {
		for rel := range fs.changes {
			fs.changeIndex.Add(rel)
		}
	}
	if fs.deleteIndex.Empty() && len(fs.tombstone) > 0 {
		for rel := range fs.tombstone {
			fs.deleteIndex.MarkDeleted(rel)
		}
	}
}

func (fs *FS) setChange(rel string, kind changeKind) {
	fs.changes[rel] = kind
	fs.changeIndex.Add(rel)
}

func (fs *FS) ensureOverlayDir(path string) error {
	clean := filepath.Clean(path)
	if _, ok := fs.overlayDirs[clean]; ok {
		return nil
	}
	if err := os.MkdirAll(clean, 0o755); err != nil {
		return err
	}
	fs.overlayDirs[clean] = struct{}{}
	return nil
}

type pathTrie struct {
	root pathTrieNode
}

type pathTrieNode struct {
	children map[string]*pathTrieNode
	terminal bool
}

func (t *pathTrie) Empty() bool {
	return len(t.root.children) == 0 && !t.root.terminal
}

func (t *pathTrie) Add(path string) {
	node := &t.root
	for _, part := range splitPath(path) {
		if node.children == nil {
			node.children = map[string]*pathTrieNode{}
		}
		next, ok := node.children[part]
		if !ok {
			next = &pathTrieNode{}
			node.children[part] = next
		}
		node = next
	}
	node.terminal = true
}

func (t *pathTrie) Remove(path string) {
	parts := splitPath(path)
	t.removeRecursive(&t.root, parts, 0)
}

func (t *pathTrie) removeRecursive(node *pathTrieNode, parts []string, index int) bool {
	if node == nil {
		return false
	}
	if index == len(parts) {
		node.terminal = false
		return len(node.children) == 0
	}
	child, ok := node.children[parts[index]]
	if !ok {
		return false
	}
	if t.removeRecursive(child, parts, index+1) {
		delete(node.children, parts[index])
	}
	return len(node.children) == 0 && !node.terminal
}

func (t *pathTrie) HasPrefix(path string) bool {
	node := &t.root
	if node.terminal {
		return true
	}
	for _, part := range splitPath(path) {
		next, ok := node.children[part]
		if !ok {
			return false
		}
		node = next
		if node.terminal {
			return true
		}
	}
	return false
}

func (t *pathTrie) MarkDeleted(path string) {
	node := &t.root
	for _, part := range splitPath(path) {
		if node.children == nil {
			node.children = map[string]*pathTrieNode{}
		}
		next, ok := node.children[part]
		if !ok {
			next = &pathTrieNode{}
			node.children[part] = next
		}
		node = next
	}
	node.terminal = true
	node.children = nil
}

func (t *pathTrie) Descendants(path string) []string {
	parts := splitPath(path)
	node := &t.root
	for _, part := range parts {
		next, ok := node.children[part]
		if !ok {
			return nil
		}
		node = next
	}
	out := make([]string, 0, 4)
	prefix := strings.Join(parts, string(os.PathSeparator))
	t.collectDescendants(node, prefix, &out)
	return out
}

func (t *pathTrie) collectDescendants(node *pathTrieNode, current string, out *[]string) {
	if node == nil {
		return
	}
	if node.terminal {
		*out = append(*out, current)
	}
	for part, child := range node.children {
		next := part
		if current != "" {
			next = current + string(os.PathSeparator) + part
		}
		t.collectDescendants(child, next, out)
	}
}

func splitPath(path string) []string {
	if path == "" {
		return nil
	}
	return strings.Split(path, string(os.PathSeparator))
}

func replaceFileStreaming(source, target string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".bus-tx-*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer func() { _ = os.Remove(tmpPath) }()
	if err := copyIntoFile(source, tmp); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, target)
}

func copyFileStreaming(source, target string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func copyIntoFile(source string, out *os.File) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
