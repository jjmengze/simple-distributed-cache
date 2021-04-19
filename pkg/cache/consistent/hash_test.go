package consistent

import (
	"k8s.io/klog/v2"
	"reflect"
	"strconv"
	"testing"
)

func mockHashFun(data []byte) uint32 {
	key, err := strconv.Atoi(string(data))
	if err != nil {
		klog.Errorf("convert key error: %v ", err)
	}
	return uint32(key)
}
func TestConsistentMap_Add(t *testing.T) {
	type fields struct {
		hashFn   HashFn
		replicas int
		keys     []int
		hashMap  map[int]string
	}
	type args struct {
		realNodes []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wants  map[int]string
	}{
		{
			name: "mock hash function",
			fields: fields{
				hashFn:   mockHashFun,
				replicas: 3,
				keys:     make([]int, 0),
				hashMap:  make(map[int]string, 0),
			},
			args: args{
				realNodes: []string{"1", "2", "3"},
			},
			wants: map[int]string{
				1: "1",
				2: "2",
				3: "3",

				11: "1",
				12: "2",
				13: "3",

				21: "1",
				22: "2",
				23: "3",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &ConsistentMap{
				hashFn:   tt.fields.hashFn,
				replicas: tt.fields.replicas,
				keys:     tt.fields.keys,
				hashMap:  tt.fields.hashMap,
			}
			cm.Add(tt.args.realNodes...)
			if !reflect.DeepEqual(cm.hashMap, tt.wants) {
				t.Errorf("want result: %v but get %v ", tt.wants, cm.hashMap)
			}
		})
	}
}

func TestConsistentMap_Get(t *testing.T) {
	type fields struct {
		hashFn   HashFn
		replicas int
		keys     []int
		hashMap  map[int]string
	}
	type args struct {
		key []string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "mock fake ConsistentMap data ",
			fields: fields{
				hashFn:   mockHashFun,
				replicas: 3,
				keys:     []int{1, 2, 3},
				hashMap: map[int]string{
					1: "1",
					2: "2",
					3: "3",

					11: "1",
					12: "2",
					13: "3",

					21: "1",
					22: "2",
					23: "3",
				},
			},
			args: args{
				key: []string{
					"1",
					"2",
					"3",
					"4",
				},
			},
			want: []string{
				"1",
				"2",
				"3",
				"1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cm := &ConsistentMap{
				hashFn:   tt.fields.hashFn,
				replicas: tt.fields.replicas,
				keys:     tt.fields.keys,
				hashMap:  tt.fields.hashMap,
			}
			for index, key := range tt.args.key {
				if got := cm.Get(key); got != tt.want[index] {
					t.Errorf("Get() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
