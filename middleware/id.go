package middleware

import (
	"sync"
	"math"
)

type IdGenerator interface {
	GetUint32() uint32 //获得一个int32类型的
}

type myIdGenerator struct {
	sn    uint32     //当前的id
	ended bool       //钱一个id是否已经为其类型所能表示的最大值
	mutex sync.Mutex //互斥锁
}


func NewIdGenerator () IdGenerator {
	return &myIdGenerator{}
}

func (gen *myIdGenerator) GetUint32() uint32 {
	gen.mutex.Lock()
	defer gen.mutex.Unlock()
	if gen.ended {
		defer func() { gen.ended = false}()
		gen.sn = 0
		return gen.sn
	}
	id := gen.sn
	if id < math.MaxUint32 {
		gen.sn++
	}else{
		gen.ended = true
	}
	return id
}
