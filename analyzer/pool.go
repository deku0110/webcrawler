package analyzer

import (
	"webcrawler/middleware"
	"reflect"
	"fmt"
	"errors"
)
type GenAnalyzer func() Analyzer
type AnalyzerPool interface {
	Take() (Analyzer, error)
	Return(analyzer Analyzer) error
	Total() uint32
	Used() uint32
}

type myAnalyzerPool struct {
	pool middleware.Pool//实体池
	etype reflect.Type
}

func NewAnalyzerPool(total uint32,gen GenAnalyzer) (AnalyzerPool,error) {
	etype := reflect.TypeOf(gen())
	genEntity := func() middleware.Entity{
		return gen()
	}
	pool, err := middleware.NewPool(total,etype,genEntity)
	if err != nil {
		return nil ,err
	}
	analyzerPool := &myAnalyzerPool{pool:pool,etype:etype}
	return analyzerPool,nil
}


func (anaPool *myAnalyzerPool) Take() (Analyzer, error) {
	entity, err := anaPool.pool.Take()
	if err != nil {
		return nil,err
	}
	ana,ok := entity.(Analyzer)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", anaPool.etype)
		panic(errors.New(errMsg))
	}
	return ana,nil
}

func (anaPool *myAnalyzerPool) Return(analyzer Analyzer) error {
	return anaPool.pool.Return(analyzer)
}

func (anaPool *myAnalyzerPool) Total() uint32 {
	return anaPool.pool.Total()
}

func (anaPool *myAnalyzerPool) Used() uint32 {
	return anaPool.pool.Used()
}
