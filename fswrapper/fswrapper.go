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
}

func NewFSWrapper(wraps fs.FS, appends string) *FSWrapper {
	return &FSWrapper{
		wrapped: wraps,
		toAdd:   appends,
	}
}

func (fs *FSWrapper) Open(name string) (fs.File, error) {
	logrus.WithFields(logrus.Fields{
		"prefix":   fs.toAdd,
		"filename": name,
		"wrapped":  fs.wrapped,
	}).Debugln("fswrapper: Opening file with prefix")
	res, err := fs.wrapped.Open(fs.toAdd + name)
	logrus.WithFields(logrus.Fields{
		"file":      res,
		"err":       err,
		"full-file": fs.toAdd + name,
	}).Debugln("fswrapper: Result of opening file")
	return res, err
}
