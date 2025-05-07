package filesystem

import (
	"os/user"
	"path/filepath"
	"strings"
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
	path = filepath.Clean(path)
	out := filepath.Join("/", path)
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
