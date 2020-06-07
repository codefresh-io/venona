package config

import (
	"os"
	"time"
)

type (
	info struct {
		name    string
		size    int64
		mode    os.FileMode
		modTime time.Time
		isDir   bool
		sys     interface{}
		// Name() string       // base name of the file
		// Size() int64        // length in bytes for regular files; system-dependent for others
		// Mode() FileMode     // file mode bits
		// ModTime() time.Time // modification time
		// IsDir() bool        // abbreviation for Mode().IsDir()
		// Sys() interface{}   // underlying data source (can return nil)
	}
)

func (i info) Name() string {
	return i.name
}

func (i info) Size() int64 {
	return i.size
}

func (i info) Mode() os.FileMode {
	return i.mode
}

func (i info) ModTime() time.Time {
	return i.modTime
}

func (i info) IsDir() bool {
	return i.isDir
}

func (i info) Sys() interface{} {
	return i.sys
}
