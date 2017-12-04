package downloader

import (
	"webcrawler/middleware"
	"reflect"
	"fmt"
	"errors"
)

type GenPageDownloader func() PageDownloader

type PageDownloaderPool interface {
	Take() (PageDownloader, error)
	Return(dl PageDownloader) error
	Total() uint32 //池的总容量
	Used() uint32  //正在被使用的网页下载器的数量
}

//网页下载器池的实现类型
type myDownloaderPool struct {
	pool  middleware.Pool //实体池
	etype reflect.Type    //池内实体的类型
}

func NewPageDownloaderPool(total uint32, gen GenPageDownloader) (PageDownloaderPool, error) {

	etype := reflect.TypeOf(gen())
	genEntity := func() middleware.Entity {
		return gen()
	}
	pool, err := middleware.NewPool(total, etype, genEntity)
	if err != nil {
		return nil, err
	}
	dlpool := &myDownloaderPool{pool: pool, etype: etype}
	return dlpool, nil
}

func (dlpool *myDownloaderPool) Take() (PageDownloader, error) {
	entity, err := dlpool.pool.Take()
	if err != nil {
		return nil,err
	}
	dl,ok := entity.(PageDownloader)
	if !ok {
		errMsg := fmt.Sprintf("The type of entity is NOT %s!\n", dlpool.etype)
		panic(errors.New(errMsg))
	}
	return dl,nil
}

func (dlpool *myDownloaderPool) Return(dl PageDownloader) error {
	return dlpool.pool.Return(dl)
}

func (dlpool *myDownloaderPool) Total() uint32 {
	return dlpool.pool.Total()
}

func (dlpool *myDownloaderPool) Used() uint32 {
	return dlpool.pool.Used()
}
