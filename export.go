package gotool

import "github.com/llyb120/gotool/collection"

func NewMap[K comparable, V any]() collection.Map[K, V] {
	return collection.NewMap[K, V]()
}
