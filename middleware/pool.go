package middleware

import (
	"reflect"
	"errors"
	"fmt"
	"sync"
)

type Pool interface {
	Take() (Entity, error)
	Return(entity Entity) error
	Total() uint32
	Used() uint32
}

type Entity interface {
	Id() uint32
}

type myPool struct {
	total       uint32
	etyep       reflect.Type
	genEntity   func() Entity
	container   chan Entity
	idContainer map[uint32]bool //实体ID的容器
	mutex       sync.Mutex      //互斥锁
}

func NewPool(total uint32, etype reflect.Type, genEntity func() Entity) (Pool, error) {
	if total == 0 {
		errMsg := fmt.Sprintf("The pool can not be initialized! (total=%d)\n", total)
		return nil, errors.New(errMsg)
	}
	size := int(total)
	idContainer := make(map[uint32]bool)
	container := make(chan Entity, size)
	for i := 0; i < size; i++ {
		newEntity := genEntity()
		if etype != reflect.TypeOf(newEntity) {
			errMsg := fmt.Sprintf("The type of result of function genEntity() is not %s!\n", etype)
			return nil, errors.New(errMsg)
		}
		container <- newEntity
		idContainer[newEntity.Id()] = true
	}
	pool := &myPool{
		total:       total,
		genEntity:   genEntity,
		container:   container,
		etyep:       etype,
		idContainer: idContainer}
	return pool, nil
}

func (pool *myPool) Take() (Entity, error) {
	entity, ok := <-pool.container
	if !ok {
		return nil, errors.New("The inner container is invalid")
	}
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	pool.idContainer[entity.Id()] = false
	return entity, nil
}

func (pool *myPool) Return(entity Entity) error {
	if entity == nil {
		return errors.New("the returnning entity is invalid!")
	}
	if pool.etyep != reflect.TypeOf(entity) {
		errMsg := fmt.Sprintf("The type of returning entity is NOT %s!\n", pool.etyep)
		return errors.New(errMsg)
	}
	entityId := entity.Id()
	result := pool.compareAndSetForIdContainer(entityId, false, true)
	if result == -1 {
		errMsg := fmt.Sprintf("The entity (id=%d) is invalid!\n", entityId)
		return errors.New(errMsg)
	}
	if result == 0 {
		errMsg := fmt.Sprintf("The entity (id=%d) is already in the pool !\n", entityId)
		return errors.New(errMsg)
}
	pool.container <- entity
	return nil
}
//结果值 -1表示键值不存在,0表示失败 1表示成功
func (pool *myPool) compareAndSetForIdContainer(entityId uint32, oldValue bool, newValue bool) uint8 {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()
	v, ok := pool.idContainer[entityId]
	if !ok {
		return -1
	}
	if v != oldValue {
		return 0
	}
	pool.idContainer[entityId] = newValue
	return 1
}

func (pool *myPool) Total() uint32 {
	return pool.total
}

func (pool *myPool) Used() uint32 {
	return pool.total - uint32(len(pool.container))
}
