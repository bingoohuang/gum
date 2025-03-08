package spin

import (
	"time"

	"github.com/charmbracelet/gum/bingoo"
)

func Timeout(timeout time.Duration) func(*Options) {
	return func(o *Options) { o.Timeout = timeout }
}

func AnyKey(anyKey bool) func(*Options) {
	return func(o *Options) { o.AnyKey = anyKey }
}

func ClearView(clearView bool) func(*Options) {
	return func(o *Options) { o.ClearView = clearView }
}

func Spin(optionsFn ...func(*Options)) error {
	option := &Options{}
	bingoo.KongParse(option, bingoo.KongVars)

	for _, fn := range optionsFn {
		fn(option)
	}

	return option.RunBingoo()
}

func (o Options) Run() error {
	return o.RunBingoo()
}
