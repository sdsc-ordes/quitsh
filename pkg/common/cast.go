package common

import (
	"reflect"

	"github.com/sdsc-ordes/quitsh/pkg/log"
)

func Cast[T any](settings any) T {
	t, ok := settings.(T)
	if !ok {
		log.Panic(
			"Wrong cast.",
			"from",
			reflect.TypeOf(settings).String(),
			"to",
			reflect.TypeFor[T]().String(),
		)
	}

	return t
}
