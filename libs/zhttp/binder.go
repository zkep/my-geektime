package zhttp

type (
	Binder interface {
		Name() string
		Binding(src any, dst any) error
	}
)
