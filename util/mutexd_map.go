package util

import "sync"

type MutexMap[X comparable, Y any] struct {
	sync.RWMutex
	Map map[X]Y
}

func NewMutexMap[X comparable, Y any]() MutexMap[X, Y] {
	return MutexMap[X, Y]{
		Map: map[X]Y{},
	}
}
