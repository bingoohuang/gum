package choose

import (
	"strings"
	"time"

	"github.com/charmbracelet/gum/bingoo"
	"github.com/charmbracelet/gum/internal/tty"
)

func Timeout(timeout time.Duration) func(*Options) {
	return func(o *Options) { o.Timeout = timeout }
}

func Limit(limit int) func(*Options) {
	return func(o *Options) { o.Limit = limit }
}
func Header(header string) func(*Options) {
	return func(o *Options) { o.Header = header }
}

func Choose(options []string, optionsFn ...func(*Options)) ([]int, []string, error) {
	option := &Options{}
	bingoo.KongParse(option, bingoo.KongVars)

	option.Options = options
	for _, fn := range optionsFn {
		fn(option)
	}
	return option.RunBingoo()
}

// Run provides a shell script interface for choosing between different through
// options.
func (o Options) Run() error {
	_, out, err := o.RunBingoo()
	if err != nil {
		return err
	}

	tty.Println(strings.Join(out, o.OutputDelimiter))
	return nil
}
