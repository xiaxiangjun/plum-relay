package utils

import "sync"

type SyncMap[K comparable, V any] struct {
	locker sync.Mutex
	data   map[K]V
}

// 存储数据
func (self *SyncMap[K, V]) Store(k K, v V) {
	self.locker.Lock()
	defer self.locker.Unlock()

	// 创建新的变量
	if nil == self.data {
		self.data = make(map[K]V)
	}

	self.data[k] = v
}

// 加载变量
func (self *SyncMap[K, V]) Load(k K) (V, bool) {
	self.locker.Lock()
	defer self.locker.Unlock()

	// 创建新的变量
	if nil == self.data {
		self.data = make(map[K]V)
	}

	v, ok := self.data[k]
	return v, ok
}

// 删除
func (self *SyncMap[K, V]) Delete(k K) {
	self.locker.Lock()
	defer self.locker.Unlock()

	// 创建新的变量
	if nil == self.data {
		self.data = make(map[K]V)
	}

	delete(self.data, k)
}

// 执行任务
func (self *SyncMap[K, V]) CompareDelete(k K, c func(V) bool) {
	self.locker.Lock()
	defer self.locker.Unlock()

	// 创建新的变量
	if nil == self.data {
		self.data = make(map[K]V)
	}

	v, ok := self.data[k]
	// 比较并删除
	if ok && c(v) {
		delete(self.data, k)
	}
}
