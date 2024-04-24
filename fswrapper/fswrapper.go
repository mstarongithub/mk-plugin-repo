package fswrapper

import (
	"io/fs"

	"github.com/sirupsen/logrus"
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
	if fs.log {
		logrus.WithFields(logrus.Fields{
			"prefix":   fs.toAdd,
			"filename": name,
			"wrapped":  fs.wrapped,
		}).Debugln("fswrapper: Opening file with prefix")
	}
	res, err := fs.wrapped.Open(fs.toAdd + name)
	if fs.log {
		logrus.WithFields(logrus.Fields{
			"file":      res,
			"err":       err,
			"full-file": fs.toAdd + name,
		}).Debugln("fswrapper: Result of opening file")
	}
	return res, err
}
