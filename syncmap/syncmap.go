package syncmap

import (
	"fmt"
	"reflect"
	"sync"
)

// SyncMap goroutine 安全的 Map 类型
type SyncMap struct {
	mutex  sync.Mutex
	values map[interface{}]interface{}
}

// New 创建 SyncMap 对象
func New() *SyncMap {
	return &SyncMap{
		values: make(map[interface{}]interface{}),
	}
}

// Get 获取元素值，若无 key 值元素，且 pvalue 指向的值不为 nil，则将 pvalue 指向的值插入到 key 对应的元素中
func (m *SyncMap) Get(key interface{}, pvalue interface{}) (has bool, err error) {
	if reflect.TypeOf(pvalue).Kind() != reflect.Ptr {
		return false, fmt.Errorf("pvalue is not pointer")
	}

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, ok := m.values[key]; !ok {
		// 根据 pvalue 所指向的内容是否为 nil 判断是否做插入动作
		if !reflect.ValueOf(pvalue).Elem().IsNil() {
			// log.Println("insert ", reflect.ValueOf(pvalue).Elem().Interface(), "to key", key)
			m.values[key] = reflect.ValueOf(pvalue).Elem().Interface()
		}

		return false, nil
	}

	reflect.ValueOf(pvalue).Elem().Set(reflect.ValueOf(m.values[key]))
	return true, nil
}

// Remove 删除元素
func (m *SyncMap) Remove(key interface{}) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	delete(m.values, key)
	return nil
}
