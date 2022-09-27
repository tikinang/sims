package list

import "container/list"

type LinkedList[T any] struct {
	l *list.List
}

func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{l: list.New()}
}

func (r *LinkedList[T]) Len() int {
	return r.l.Len()
}

func (r *LinkedList[T]) IterateRemove(iterator func(T) bool) {
	for e := r.l.Front(); e != nil; e = e.Next() {
		if iterator(e.Value.(T)) {
			r.l.Remove(e)
		}
	}
}

func (r *LinkedList[T]) Iterate(iterator func(T)) {
	for e := r.l.Front(); e != nil; e = e.Next() {
		iterator(e.Value.(T))
	}
}

func (r *LinkedList[T]) PushBack(v T) {
	r.l.PushBack(v)
}
