package lru

import (
	"container/list"
	"log"
	"testing"
)

func mockOnEvictedFn(key string, value interface{}) error {
	log.Print("Evicted KEY:", key)
	log.Print("Evicted VALUE:", value)
	return nil
}

func mockNewLRUCache(options *Options) lru {
	return lru{
		data:           make(map[string]*list.Element),
		maxElementSize: options.maxElementSize,
		maxBytes:       options.maxBytes,
		totalBytes:     0,
		OnEvicted:      options.fn,
		list:           list.New(),
	}
}

func Test_lru_Set(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name      string
		fields    lru
		args      args
		wantValue interface{}
		wantErr   bool
	}{
		{
			name:   "LRU set fun function with default options",
			fields: mockNewLRUCache(DefaultOptions(mockOnEvictedFn)),
			args: args{
				key:   "foo",
				value: "bar",
			},
			wantValue: "bar",
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lru{
				data:           tt.fields.data,
				maxElementSize: tt.fields.maxElementSize,
				maxBytes:       tt.fields.maxBytes,
				totalBytes:     tt.fields.totalBytes,
				OnEvicted:      tt.fields.OnEvicted,
				list:           tt.fields.list,
			}
			if err := l.Set(tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("Set() error = %v, wantErr %v", err, tt.wantErr)
			}

			if element := l.data[tt.args.key].Value; element.(*data).val != tt.wantValue {
				t.Errorf("Want to Set() cache %v value but get %v", tt.wantValue, element.(*data).val)
			}
		})
	}
}
