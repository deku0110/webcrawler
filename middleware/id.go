package middleware

type IdGenerator interface {
	GetUint32() uint32//获得一个int32类型的
}
