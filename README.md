# Memphis

Status: Work in Progress

Memphis is a virtual (memory) file system for golang. It provides the same functionality (and is meant to be used as) a backing store for [billy](https://github.com/go-git/go-billy), [rio](https://github.com/polydawn/rio), and others.

Memphis stores can also be generated from on-disk directory trees. File contents of unmodified files will be read from disk, while write requests to a file will transition them to in-memory content buffers.
