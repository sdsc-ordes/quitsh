package config

type Initer interface {
	Init() error
}

// Init runs [Initer.Init] if the type `v` fulfills the interface [Initer].
// TODO: Maybe use a reflect iteration?
func Init(v any) error {
	initable, ok := v.(Initer)
	if ok {
		return initable.Init()
	}

	return nil
}
