package filesystem

import (
	"testing"
)

func TestAbsolutePath(t *testing.T) {
	type testcase struct {
		root, path, expected string
	}
	testCases := []testcase{
		{root: "/", path: "/", expected: "/"},
		{root: "/home/sy", path: "/", expected: "/home/sy"},
		{root: "/", path: "~/", expected: "/home/sy"},
		{root: "/", path: "~/..", expected: "/home"},
		{root: "/home/sy", path: "documents", expected: "/home/sy/documents"},
		{root: "/home/sy", path: "/documents", expected: "/home/sy/documents"},
		{root: "/home/sy/test", path: "~/", expected: "/home/sy/test"},
		{root: "/home/sy", path: "../", expected: "/home/sy"},
		{root: "/home/sy", path: "documents/../../pictures", expected: "/home/sy/pictures"},
	}
	for _, tc := range testCases {
		actual := AbsolutePath(tc.root, tc.path)
		if actual != tc.expected {
			actual := AbsolutePath(tc.root, tc.path)
			t.Errorf("Expected %s for %s and %s, got %s", tc.expected, tc.root, tc.path, actual)
		} else {
			t.Log(tc.root, "and", tc.path, "resolved to", tc.expected)
		}
	}
}

func TestRelativePath(t *testing.T) {
	type testcase struct {
		root, path, expected string
	}
	testCases := []testcase{
		{root: "/", path: "/", expected: "/"},
		{root: "/home/sy", path: "/home/sy", expected: "/"},
		{root: "/", path: "/home/sy", expected: "/home/sy"},
		{root: "/home/sy", path: "/home/sy/documents", expected: "/documents"},
	}
	for _, tc := range testCases {
		actual := RelativePath(tc.root, tc.path)
		if actual != tc.expected {
			actual := RelativePath(tc.root, tc.path)
			t.Errorf("Expected %s for %s and %s, got %s", tc.expected, tc.root, tc.path, actual)
		} else {
			t.Log(tc.root, "and", tc.path, "resolved to", tc.expected)
		}
	}
}
