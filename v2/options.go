package excelizex

// EachOption configures Each/EachMap behavior.
type EachOption func(*readOptions)

type readOptions struct {
	concurrency int
	failFast    bool
}

// Concurrency sets worker count for Each/EachMap (default 1).
func Concurrency(n int) EachOption {
	return func(o *readOptions) {
		if n < 1 {
			n = 1
		}
		o.concurrency = n
	}
}

// FailFast aborts on the first bind/validate/callback error.
func FailFast() EachOption {
	return func(o *readOptions) {
		o.failFast = true
	}
}

func applyReadOptions(opts []EachOption) readOptions {
	cfg := readOptions{concurrency: 1}
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.concurrency < 1 {
		cfg.concurrency = 1
	}

	return cfg
}
