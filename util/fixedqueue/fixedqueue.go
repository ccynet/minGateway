// 一个线程安全的固定长度队列，压入超过队列长度的数据时，替换最老的数据
// 使用环形队列的方式存储数据，减少内存的创建开销和碎片产生
// 经对比测试，比直接使用slice增删方式实现的队列性能好一半左右（测试的一组数据是：3560 ns/op，对比数据为：8876 ns/op）
// Email:cgrencn@gmail.com

package fixedqueue

import (
	"fmt"
	"sync"
)

type FixedQueue struct {
	mu    *sync.Mutex
	list  []interface{}
	size  int  //队列大小
	p1    int  //队列头位置
	p2    int  //队列尾位置
	empty bool //是否列表为空
}

func NewFixedQueue(size int) *FixedQueue {
	q := new(FixedQueue)
	q.mu = new(sync.Mutex)
	q.list = make([]interface{}, size)
	q.size = size
	q.empty = true
	//q.p1 = 0
	//q.p2 = 0
	return q
}

//压入队列尾一个数据
func (q *FixedQueue) Push(data interface{}) {
	q.mu.Lock()

	if q.empty {
		q.empty = false
		q.list[q.p2] = data
		q.mu.Unlock()
		return
	}

	if q.p2+1 == q.size {
		q.p2 = 0
	} else {
		q.p2++
	}

	q.list[q.p2] = data

	if q.p1 == q.p2 {
		//追上头，吃掉最老的数据
		q.p1++
		if q.p1 == q.size {
			q.p1 = 0
		}
	}

	q.mu.Unlock()
	return
}

//从队列头取出一个数据，同时从队列中删除它
func (q *FixedQueue) Pop() (interface{}, bool) {
	q.mu.Lock()

	if q.empty {
		q.mu.Unlock()
		return nil, false
	}

	data := q.list[q.p1]

	if q.p1 == q.p2 {
		//取出最后一个数据
		q.empty = true
		q.mu.Unlock()
		return data, true
	}

	if q.p1+1 == q.size {
		q.p1 = 0
	} else {
		q.p1++
	}

	q.mu.Unlock()
	return data, true
}

//获取队列头的数据，但不会从队列中删除
func (q *FixedQueue) Get() (interface{}, bool) {
	q.mu.Lock()

	if q.empty {
		q.mu.Unlock()
		return nil, false
	}

	data := q.list[q.p1]

	q.mu.Unlock()
	return data, true
}

//获取队列的长度
func (q *FixedQueue) Len() int {
	q.mu.Lock()
	if q.empty {
		q.mu.Unlock()
		return 0
	}
	l := q.p2 - q.p1 + 1
	if l <= 0 {
		l = q.size + l
	}
	q.mu.Unlock()
	return l
}

//清空队列，使队列长度为零
func (q *FixedQueue) Clear() {
	q.mu.Lock()
	q.p1 = 0
	q.p2 = 0
	q.empty = true
	q.mu.Unlock()
}

func (q *FixedQueue) Print() {
	q.mu.Lock()
	fmt.Println(q.list)
	q.mu.Unlock()
}
