package lru

import (
	"container/list"
	"simple-distributed-cache/pkg/cache"
	"sync"
)

type OnEvictedFn func(key string, value interface{}) error

type lru struct {
	mu             sync.RWMutex
	data           map[string]*list.Element
	maxElementSize int64
	maxBytes       int64
	totalBytes     int64
	OnEvicted      func(key string, value interface{}) error
	list           *list.List
}

func NewLRUWithOption(options ...Option) cache.SetterGetter {
	opts := &Options{}
	for _, opt := range options {
		if opt != nil {
			if err := opt(opts); err != nil {
				return nil
			}
		}
	}
	return NewLRUCache(opts)
}
func NewLRUCache(options *Options) cache.SetterGetter {
	return &lru{
		data:           make(map[string]*list.Element),
		maxElementSize: options.maxElementSize,
		maxBytes:       options.maxBytes,
		totalBytes:     0,
		OnEvicted:      options.fn,
		list:           list.New(),
	}
}

type data struct {
	val interface{}
	key interface{}
}

func (l *lru) Set(key, value string) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if exitElement, ok := l.data[key]; ok {
		l.list.MoveToFront(exitElement)
		exitElement.Value = value
		return nil
	}
	element := l.list.PushFront(&data{key: key, val: value})
	l.data[key] = element
	if int64(len(l.data)) > l.maxElementSize || l.totalBytes > l.maxBytes {
		element := l.list.Back()
		err := l.OnEvicted(key, element.Value)
		if err != nil {
			return err
		}
		l.list.Remove(element)
		delete(l.data, key)
	}
	return nil
}

func (l lru) Get(key string) (string, error) {
	panic("implement me")
}

// Option is a function on the options for a lru setting.
type Option func(*Options) error

// Options can be used to create a customized lru setting.
type Options struct {
	maxElementSize int64
	maxBytes       int64
	fn             OnEvictedFn
}

func MaxElementSize(maxElementSize int64) Option {
	return func(o *Options) error {
		o.maxElementSize = maxElementSize
		return nil
	}
}

func MaxBytes(maxBytes int64) Option {
	return func(o *Options) error {
		o.maxBytes = maxBytes
		return nil
	}
}

func OnEvicted(fn OnEvictedFn) Option {
	return func(o *Options) error {
		o.fn = fn
		return nil
	}
}

func DefaultOptions(fn OnEvictedFn) *Options {
	return &Options{
		maxElementSize: 100,
		maxBytes:       100,
		fn:             fn,
	}
}
