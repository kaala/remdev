package handler

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/net/webdav"
)

// NewWebDAV creates a WebDAV handler restricted to the given root directory.
func NewWebDAV(root string) (http.Handler, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	fs := &rootFS{
		root: absRoot,
		fs:   webdav.Dir(absRoot),
	}

	handler := &webdav.Handler{
		Prefix:     "/dav",
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				println("webdav:", r.Method, r.URL.Path, err.Error())
			}
		},
	}

	return handler, nil
}

// rootFS wraps a webdav.Dir to enforce root path restrictions.
type rootFS struct {
	root string
	fs   webdav.Dir
}

func (f *rootFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if !f.isSafe(name) {
		return os.ErrPermission
	}
	return f.fs.Mkdir(ctx, name, perm)
}

func (f *rootFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	if !f.isSafe(name) {
		return nil, os.ErrPermission
	}
	return f.fs.OpenFile(ctx, name, flag, perm)
}

func (f *rootFS) RemoveAll(ctx context.Context, name string) error {
	if !f.isSafe(name) {
		return os.ErrPermission
	}
	return f.fs.RemoveAll(ctx, name)
}

func (f *rootFS) Rename(ctx context.Context, oldName, newName string) error {
	if !f.isSafe(oldName) || !f.isSafe(newName) {
		return os.ErrPermission
	}
	return f.fs.Rename(ctx, oldName, newName)
}

func (f *rootFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	if !f.isSafe(name) {
		return nil, os.ErrPermission
	}
	return f.fs.Stat(ctx, name)
}

func (f *rootFS) isSafe(name string) bool {
	resolved := filepath.Clean(filepath.Join(f.root, name))
	return strings.HasPrefix(resolved, f.root)
}
