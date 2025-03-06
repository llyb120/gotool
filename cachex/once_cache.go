package cachex

import (
	"context"
	"sync"
	"time"
)

// 一次性缓存，超过多久即会销毁

type OnceCache[T any] struct {
	mu    sync.RWMutex
	cache map[string]cacheItemWrapper[T]
	opts  OnceCacheOption
}

type OnceCacheOption struct {
	Expire           time.Duration
	DefaultKeyExpire time.Duration
	CheckInterval    time.Duration
	Destroy          func()
}

func NewOnceCache[T any](opts OnceCacheOption) *OnceCache[T] {
	cache := &OnceCache[T]{
		opts:  opts,
		cache: make(map[string]cacheItemWrapper[T]),
	}
	go cache.start()
	return cache
}

func (c *OnceCache[T]) start() {
	if c.opts.Destroy != nil {
		defer c.opts.Destroy()
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.opts.Expire)
	defer cancel()

	if c.opts.CheckInterval > 0 {
		// 小于等于0的时候永不过期
		go func() {
			ticker := time.NewTicker(c.opts.CheckInterval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					func() {
						c.mu.Lock()
						// 执行检查操作
						defer c.mu.Unlock()
						mp := make(map[string]cacheItemWrapper[T])
						now := time.Now()
						for key, item := range c.cache {
							if item.canExpire && item.expire.After(now) {
								mp[key] = item
							}
						}
						c.cache = mp
					}()
				}
			}
		}()
	}

	<-ctx.Done()
}

func (c *OnceCache[T]) Set(key string, value T) {
	c.SetExpire(key, value, c.opts.DefaultKeyExpire)
}

func (c *OnceCache[T]) SetExpire(key string, value T, expire time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = cacheItemWrapper[T]{
		value:     value,
		expire:    time.Now().Add(expire),
		canExpire: expire > 0,
	}
}

func (c *OnceCache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.cache[key]
	if !ok {
		return item.value, false
	}
	return item.value, true
}

func (c *OnceCache[T]) Del(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, key)
}

func (c *OnceCache[T]) GetOrSetFunc(key string, fn func() T) T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.cache[key]
	if !ok {
		value := fn()
		c.cache[key] = cacheItemWrapper[T]{
			value:  value,
			expire: time.Now().Add(c.opts.DefaultKeyExpire),
		}
		return value
	}
	return item.value
}
