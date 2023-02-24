package consistenthash

import (
	"hash/crc32"
	"sort"
)

type HashFunc func(data []byte) uint32

type NodeMap struct {
	hashFunc    HashFunc       // 可以传不同的hash函数
	nodeHashs   []int          // 需要排序  12343， 89765
	nodehashMap map[int]string // string 是记录的节点信息
}

func NewNodeMap(fn HashFunc) *NodeMap {
	m := &NodeMap{
		hashFunc:    fn,
		nodehashMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

// 判断整个集群是不是空的
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodeHashs) == 0
}

// 将新的节点添加到集群中
// key: 可以是节点名称，可以是ip + port
func (m *NodeMap) AddNode(keys ...string) {
	for _, key := range keys {
		if key == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(key)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodehashMap[hash] = key
	}
	sort.Ints(m.nodeHashs)
}

// 每个key 应该去哪个节点, 选择一个节点
// key: 可以是节点名称，可以是ip + port
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}
	hash := int(m.hashFunc([]byte(key)))
	// 之前的12343  23123 58989
	// 现在新增的18989  60000
	// 直接用这个进行搜索 sort.Search， 得到的是小于等于数组里面的数字的序号，如果是最大，则需要将idx = 0
	idx := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	if idx == len(m.nodeHashs) {
		idx = 0
	}
	// idx 就是我们找的节点序号
	return m.nodehashMap[m.nodeHashs[idx]]
}
