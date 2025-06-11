package queryerslice

type options struct {
	CopySlice bool
}

// Option specify option for QueryerSlice.
type Option func(opts *options)

// WithCopy force to copy subslice for Query method.
func WithCopy() Option {
	return func(opts *options) {
		opts.CopySlice = true
	}
}
