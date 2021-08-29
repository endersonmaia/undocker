package rootfs

type options struct {
	filePrefix string
}

type Option interface {
	apply(*options)
}

type filePrefixOption string

func (p filePrefixOption) apply(opts *options) {
	opts.filePrefix = string(p)
}

// WithFilePrefixOption adds a prefix to all files in the output archive.
func WithFilePrefix(p string) Option {
	return filePrefixOption(p)
}
