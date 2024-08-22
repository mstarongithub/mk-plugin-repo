package fswrapper

import (
	"io/fs"

	"github.com/rs/zerolog/log"
)

// Fix for go:embed file systems including the full path of the embedded files
// Adds a given string to the front of all requests
type FSWrapper struct {
	wrapped fs.FS
	toAdd   string
	log     bool
}

func NewFSWrapper(wraps fs.FS, appends string, logAccess bool) *FSWrapper {
	return &FSWrapper{
		wrapped: wraps,
		toAdd:   appends,
		log:     logAccess,
	}
}

func (fs *FSWrapper) Open(name string) (fs.File, error) {
	res, err := fs.wrapped.Open(fs.toAdd + name)
	if fs.log {
		log.Debug().
			Str("prefix", fs.toAdd).
			Str("filename", name).
			Err(err).
			Msg("fswrapper: File access result")
	}
	return res, err
}
