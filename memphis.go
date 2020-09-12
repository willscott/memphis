package memphis

// Separator provides a stable path separator the memphis FS expects.
const Separator = "/"

// New creates a new, empty memphis instance
func New() *Tree {
	return newTree(0, 0, 0777)
}

// FromOS creates a memphis instance overlayed on an OS subtree
func FromOS(osPath string) *Tree {
	fs := newTree(0, 0, 0777)
	fs.deferred = deferredOSDir(fs, osPath)
	fs.ready.Do(fs.deferred)
	return fs
}
