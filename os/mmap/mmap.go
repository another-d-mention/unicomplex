package mmap

// Prot defines how the memory-mapped region is accessible
type Prot int

const (
	// PROT_NONE No access to the memory-mapped region.
	// You cannot read, write, or execute the mapped memory.
	// Useful to reserve address space or create guard pages.
	// Any access will cause a segmentation fault.
	PROT_NONE Prot = iota
	// PROT_READ Read access to the memory-mapped region.
	// Combine with MAP_SHARED to read a file without modifying it.
	PROT_READ
	// PROT_WRITE Write access to the memory-mapped region.
	// No automatically readable unless you also set PROT_READ.
	// Writing to the memory works, but reading might segfault if PROT_READ is not set on some systems.
	// Must combine with PROT_READ for read-write access in most practical cases.
	PROT_WRITE
	// PROT_EXEC Execute access to the memory-mapped region.
	// Used for mapping shared libraries (e.g., ELF binaries), JIT compilers (LLVM, V8), executable
	// stacks or code stubs.
	// Most systems restrict or require special privileges for PROT_EXEC, especially with PROT_WRITE.
	PROT_EXEC
)

// MapType defines how changes to the mapped memory affect the underlying file and other processes.
type MapType int

const (
	// MAP_SHARED Changes you make to the memory-mapped region are written back to the file and visible
	// to other processes mapping the same file with MAP_SHARED.
	// Most common for persistent changes
	MAP_SHARED MapType = 1
	// MAP_PRIVATE Creates a copy-on-write mapping. Changes are not written back to the file and are
	// not visible to other processes.
	// Safe for read-only or sandboxed writes.
	MAP_PRIVATE MapType = 2
	// MAP_SHARED_VALIDATE is like MAP_SHARED, but adds stricter checks for incompatible flag combinations.
	// Like MAP_SHARED, but stricter checks.
	MAP_SHARED_VALIDATE MapType = 3
)
