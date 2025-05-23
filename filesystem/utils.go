package filesystem

import (
	"os/user"
	"path"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func ResolveVirtualPath(root, p string) string {
	// Normalize root
	root = path.Clean("/" + strings.Trim(root, "/"))

	// Clean and join the path
	joined := path.Join("/", p) // Treat path as relative to "/"
	parts := strings.Split(joined, "/")

	// Walk and clean components, discarding ".." that would escape
	var cleanParts []string
	for _, part := range parts {
		if part == ".." {
			if len(cleanParts) > 0 {
				cleanParts = cleanParts[:len(cleanParts)-1]
			}
			// else ignore, prevents escaping
		} else if part != "" && part != "." {
			cleanParts = append(cleanParts, part)
		}
	}

	// Rebuild within root
	return path.Join(root, path.Join(cleanParts...))
}

func AbsolutePath(root, path string) string {
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

func RelativePath(root, path string) string {
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
