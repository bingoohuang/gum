package input

import (
	"fmt"
	"time"

	"github.com/charmbracelet/gum/bingoo"
)

func Timeout(timeout time.Duration) func(*Options) {
	return func(o *Options) { o.Timeout = timeout }
}

func Placeholder(placehold string) func(*Options) {
	return func(o *Options) { o.Placeholder = placehold }
}

func Prompt(prompt string) func(*Options) {
	return func(o *Options) { o.Prompt = prompt }
}

func Input(optionsFn ...func(*Options)) (string, error) {
	option := &Options{}
	bingoo.KongParse(option, bingoo.KongVars)

	for _, fn := range optionsFn {
		fn(option)
	}

	return option.RunBingoo()
}

func (o Options) Run() error {
	out, err := o.RunBingoo()
	if err != nil {
		return err
	}
	fmt.Println(out)
	return nil
}
