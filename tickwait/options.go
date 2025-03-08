package tickwait

import "time"

type Options struct {
	Timeout   time.Duration `help:"Timeout of the tick wait"`
	TimeoutFn func(cost time.Duration, view string)
	DoneFn    func(cost time.Duration, view string, result string)
}
