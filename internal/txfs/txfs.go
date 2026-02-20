package txfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
	root      string
	overlay   string
	changes   map[string]changeKind
	tombstone map[string]struct{}
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
		root:      absRoot,
		overlay:   absOverlay,
		changes:   map[string]changeKind{},
		tombstone: map[string]struct{}{},
	}, nil
}

func (fs *FS) Open(name string) (*os.File, error) {
	rel, err := cleanRel(name)
	if err != nil {
		return nil, err
	}
	if fs.isDeleted(rel) {
		return nil, os.ErrNotExist
	}
	overlayPath := filepath.Join(fs.overlay, rel)
	if exists(overlayPath) {
		return os.Open(overlayPath)
	}
	return os.Open(filepath.Join(fs.root, rel))
}

func (fs *FS) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	rel, err := cleanRel(name)
	if err != nil {
		return nil, err
	}
	overlayPath := filepath.Join(fs.overlay, rel)
	if err := os.MkdirAll(filepath.Dir(overlayPath), 0o755); err != nil {
		return nil, err
	}
	if err := fs.materializeForWrite(rel, flag, perm); err != nil {
		return nil, err
	}
	delete(fs.tombstone, rel)
	fs.changes[rel] = changeReplace
	return os.OpenFile(overlayPath, flag, perm)
}

func (fs *FS) MkdirAll(path string, perm os.FileMode) error {
	rel, err := cleanRel(path)
	if err != nil {
		return err
	}
	delete(fs.tombstone, rel)
	if err := os.MkdirAll(filepath.Join(fs.overlay, rel), perm); err != nil {
		return err
	}
	fs.changes[rel] = changeReplace
	return nil
}

func (fs *FS) Remove(name string) error {
	rel, err := cleanRel(name)
	if err != nil {
		return err
	}
	fs.markDelete(rel)
	_ = os.Remove(filepath.Join(fs.overlay, rel))
	return nil
}

func (fs *FS) RemoveAll(path string) error {
	rel, err := cleanRel(path)
	if err != nil {
		return err
	}
	fs.markDelete(rel)
	_ = os.RemoveAll(filepath.Join(fs.overlay, rel))
	return nil
}

func (fs *FS) Rename(oldPath, newPath string) error {
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
	fs.changes[newRel] = changeReplace
	return nil
}

func (fs *FS) Stat(name string) (os.FileInfo, error) {
	rel, err := cleanRel(name)
	if err != nil {
		return nil, err
	}
	if fs.isDeleted(rel) {
		return nil, os.ErrNotExist
	}
	overlayPath := filepath.Join(fs.overlay, rel)
	if info, err := os.Stat(overlayPath); err == nil {
		return info, nil
	}
	return os.Stat(filepath.Join(fs.root, rel))
}

func (fs *FS) Commit() error {
	paths := make([]string, 0, len(fs.changes))
	for rel := range fs.changes {
		paths = append(paths, rel)
	}
	sort.Strings(paths)

	for _, rel := range paths {
		kind := fs.changes[rel]
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
	overlayPath := filepath.Join(fs.overlay, rel)
	if exists(overlayPath) {
		return nil
	}
	if flag&os.O_TRUNC != 0 || flag&os.O_CREATE != 0 {
		return nil
	}
	basePath := filepath.Join(fs.root, rel)
	if !exists(basePath) {
		return nil
	}
	return copyFileStreaming(basePath, overlayPath, perm)
}

func (fs *FS) isDeleted(rel string) bool {
	for p := range fs.tombstone {
		if rel == p || strings.HasPrefix(rel, p+string(os.PathSeparator)) {
			return true
		}
	}
	return false
}

func (fs *FS) markDelete(rel string) {
	fs.tombstone[rel] = struct{}{}
	fs.changes[rel] = changeDelete
	for p := range fs.changes {
		if p != rel && strings.HasPrefix(p, rel+string(os.PathSeparator)) {
			delete(fs.changes, p)
		}
	}
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
	if err := tmp.Sync(); err != nil {
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
	return out.Sync()
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
