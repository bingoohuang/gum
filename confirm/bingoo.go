package confirm

import (
	"time"

	"github.com/charmbracelet/gum/bingoo"
	"github.com/charmbracelet/gum/internal/exit"
)

func Timeout(timeout time.Duration) func(*Options) {
	return func(o *Options) { o.Timeout = timeout }
}

func Default(defaultValue bool) func(*Options) {
	return func(o *Options) { o.Default = defaultValue }
}

func Affirmative(affirmative string) func(*Options) {
	return func(o *Options) { o.Affirmative = affirmative }
}

func Negative(negative string) func(*Options) {
	return func(o *Options) { o.Negative = negative }
}

func Prompt(prompt string) func(*Options) {
	return func(o *Options) { o.Prompt = prompt }
}

func PromptFn(promptFn func() string) func(*Options) {
	return func(o *Options) { o.PromptFn = promptFn }
}

func Confirm(optionsFn ...func(*Options)) (bool, error) {
	option := &Options{}
	bingoo.KongParse(option, bingoo.KongVars)

	for _, fn := range optionsFn {
		fn(option)
	}

	return option.RunBingoo()
}

func (o Options) Run() error {
	ok, err := o.RunBingoo()
	if err != nil {
		return err
	}
	if !ok {
		return exit.ErrExit(1)
	}

	return nil
}
