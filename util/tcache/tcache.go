// 一个线程安全的可以设置过期时间的键值对存储
// 基于性能考虑，采用被动式清除过期数据方式，即除非执行其中任一方法(Set,Get,Len,Update)，否则不会去移除过期数据
// 一般来说这些过期数据存在不会有什么影响，但如果这些数据带来问题，可在外部增加一个定时器，定时执行Update方法
// Email:cgrencn@gmail.com

package tcache

import (
	"container/list"
	"sync"
	"time"
)

type timeKey struct {
	inTime time.Time
	key    interface{}
}

type TCache struct {
	keyQueue    *list.List
	repeats     map[interface{}]int
	datas       map[interface{}]interface{}
	mu          *sync.Mutex
	destroyTime time.Duration
}

/**
destroyTime:设置过期时间
*/
func NewTimeCache(destroyTime time.Duration) *TCache {
	tc := new(TCache)
	tc.datas = make(map[interface{}]interface{}) //数据存储
	tc.repeats = make(map[interface{}]int)       //key重复出现的次数
	tc.keyQueue = list.New()                     //key存储队列
	tc.mu = new(sync.Mutex)
	tc.destroyTime = destroyTime
	return tc
}

func (tc *TCache) Set(key, data interface{}) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	//先移除过期的数据
	tc.removeOverdue()
	//设置数据
	_, exist := tc.datas[key]
	tc.datas[key] = data
	tkey := timeKey{
		inTime: time.Now(),
		key:    key,
	}
	tc.keyQueue.PushBack(tkey)
	//是否原来已经存在
	if exist {
		count, ok := tc.repeats[key] //重复了几次
		if ok {
			tc.repeats[key] = count + 1
		} else {
			tc.repeats[key] = 1
		}
	}
}

func (tc *TCache) Get(key interface{}) (interface{}, bool) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	//先移除过期的数据
	tc.removeOverdue()
	//获取数据
	d, ok := tc.datas[key]
	return d, ok
}

func (tc *TCache) Len() int {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	//先移除过期的数据
	tc.removeOverdue()
	//len
	return len(tc.datas)
}

func (tc *TCache) Update() {
	tc.mu.Lock()
	//移除过期的数据
	tc.removeOverdue()
	tc.mu.Unlock()
}

//no thread safe
func (tc *TCache) removeOverdue() {
	if tc.keyQueue.Len() == 0 {
		return
	}
	for {
		e := tc.keyQueue.Front()
		if e == nil {
			return
		}
		tkey := (e.Value).(timeKey)
		if time.Now().Sub(tkey.inTime) > tc.destroyTime {
			tc.keyQueue.Remove(e)
			if count, ok := tc.repeats[tkey.key]; !ok {
				//只有repeats里没有此key，才去删除数据
				delete(tc.datas, tkey.key)
			} else {
				if count <= 1 {
					delete(tc.repeats, tkey.key)
				} else {
					tc.repeats[tkey.key] = count - 1
				}
			}
			continue
		} else {
			break
		}
	}
}
