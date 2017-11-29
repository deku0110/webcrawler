package downloader

type PageDownloaderPool interface {
	Take() (PageDownloader, error)
	Return(dl PageDownloader) error
	Total() uint32 //池的总容量
	Used() uint32  //正在被使用的网页下载器的数量
}
