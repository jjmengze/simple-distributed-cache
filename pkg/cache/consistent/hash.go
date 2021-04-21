package consistent

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type HashFn func(data []byte) uint32

const (
	DefaultReplicas = 3
)

type ConsistentMap struct {
	hashFn   HashFn
	replicas int
	keys     []int
	hashMap  map[int]string
}

func NewConsistentMAP(replicas int, hashFn HashFn) *ConsistentMap {
	cm := &ConsistentMap{
		hashFn:   hashFn,
		replicas: replicas,
		keys:     make([]int, 0),
		hashMap:  make(map[int]string, 0),
	}
	if cm.hashFn == nil {
		cm.hashFn = crc32.ChecksumIEEE
	}
	return cm
}

func (cm *ConsistentMap) Add(realNodes ...string) {
	for _, node := range realNodes {
		for i := 0; i < cm.replicas; i++ {
			hash := int(cm.hashFn([]byte(strconv.Itoa(i) + node)))
			cm.keys = append(cm.keys, hash)
			cm.hashMap[hash] = node
		}
	}
	sort.Ints(cm.keys)
}

func (cm *ConsistentMap) Get(key string) string {
	hash := int(cm.hashFn([]byte(key)))
	index := sort.Search(len(cm.keys), func(i int) bool {
		return cm.keys[i] >= hash
	})
	return cm.hashMap[cm.keys[index%len(cm.keys)]]
}
