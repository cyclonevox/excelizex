package excelizex

// EachOption configures Each/EachMap behavior.
type EachOption func(*readOptions)

type readOptions struct {
	concurrency    int
	concurrencySet bool
	failFast       bool
}

// Concurrency sets worker count for Each/EachMap.
// An explicit Concurrency(1) overrides SetConcurrency on the builder.
// When omitted, Each uses SetConcurrency if set, otherwise 1.
func Concurrency(n int) EachOption {
	return func(o *readOptions) {
		if n < 1 {
			n = 1
		}
		o.concurrency = n
		o.concurrencySet = true
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

func resolveConcurrency(cfg readOptions, builderDefault int) int {
	if cfg.concurrencySet {
		return cfg.concurrency
	}
	if builderDefault > 1 {
		return builderDefault
	}

	return 1
}
