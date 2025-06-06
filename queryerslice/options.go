package queryerslice

type options struct {
	CopySlice bool
}

type Option func(opts *options)

func WithCopy() Option {
	return func(opts *options) {
		opts.CopySlice = true
	}
}
