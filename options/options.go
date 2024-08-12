package options

type Option[T any] interface {
	Apply(*T)
}

type FuncOption[T any] struct {
	f func(*T)
}

func (fo *FuncOption[T]) Apply(o *T) {
	if fo.f != nil {
		fo.f(o)
	}
}

func NewFuncOption[T any](f func(*T)) *FuncOption[T] {
	return &FuncOption[T]{
		f: f,
	}
}
