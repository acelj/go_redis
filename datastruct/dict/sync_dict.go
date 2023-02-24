package dict

import (
	"sync"
)

type SyncDict struct {
	m sync.Map
}

func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

func (s *SyncDict) Get(key string) (val interface{}, exists bool) {
	val, ok := s.m.Load(key)
	return val, ok
}

func (s *SyncDict) Len() int {
	lenth := 0
	s.m.Range(func(key, value interface{}) bool {
		lenth++
		return true
	})
	return lenth
}

func (s *SyncDict) Put(key string, val interface{}) (result int) {
	//panic("implement me")
	_, existed := s.m.Load(key)
	s.m.Store(key, val)
	if existed { // 这里是更新了， 所以返回的是0
		return 0
	}
	return 1
}

func (s *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	// 存在就不插入
	_, existed := s.m.Load(key)
	if existed {
		return 0
	}
	s.m.Store(key, val)
	return 1
}

func (s *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	// 存在就不插入
	_, existed := s.m.Load(key)
	if existed {
		s.m.Store(key, val)
		return 1
	}
	return 0
}

func (s *SyncDict) Remove(key string) (result int) {
	_, existed := s.m.Load(key)
	s.m.Delete(key)
	if existed {
		return 1
	}
	return 0
}

func (s *SyncDict) ForEach(consumer Consumer) {
	s.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		return true
	})
}

func (s *SyncDict) Keys() []string {
	//panic("implement me")
	result := make([]string, s.Len())
	i := 0
	s.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		return true // 只有return true 才会施加下一个
	})
	return result
}

func (s *SyncDict) RandomKeys(limit int) []string {
	// 随机返回100key, 可以重复
	result := make([]string, s.Len())
	for i := 0; i < limit; i++ {
		s.m.Range(func(key, value interface{}) bool {
			result[i] = key.(string)
			return false // 不让这个循环作用下一个
		})
	}
	return result
}

func (s *SyncDict) RandomDistinctKeys(limit int) []string {
	result := make([]string, s.Len())
	i := 0
	s.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		if i == limit {
			return false
		}
		return true
	})
	return result
}

func (s *SyncDict) Clear() {
	// 直接换一个新的，旧的可以让系统自动 gc
	*s = *MakeSyncDict()
}
