package filesystem

import (
	"os/user"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func absolutePath(root, path string) string {
	if path == "" || path == "/" {
		return root
	}

	if strings.HasPrefix(path, "~/") {
		if root == "/" {
			usr, _ := user.Current()
			dir := usr.HomeDir
			path = filepath.Join(dir, path[2:])
		} else {
			path = path[1:]
		}
	}

	if !filepath.IsAbs(path) && root == "/" {
		path, _ = filepath.Abs(path)
	}

	out := filepath.Clean(filepath.Join(root, path))
	if !strings.HasPrefix(out, root) {
		out = filepath.Clean(filepath.Join(root, filepath.Clean(path)))
	}

	path = filepath.Clean(path)
	out = filepath.Join("/", path)
	out = filepath.Join(root, out)

	if !strings.HasPrefix(out, root) {
		return root
	}

	return out
}

func relativePath(root, path string) string {
	if path == "" || path == "/" {
		return root
	}
	path = strings.TrimPrefix(path, root)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return path
}

func GetBlockSize(path string) (int, error) {
	var stat unix.Statfs_t
	err := unix.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}

	return int(stat.Bsize), nil
}

func GetPageSize() int {
	return syscall.Getpagesize()
}
