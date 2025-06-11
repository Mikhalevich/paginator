package metrics

// Noop struct for noop impl CacheMetrics.
type Noop struct {
}

// NewNoop constructs new Noop.
func NewNoop() *Noop {
	return &Noop{}
}

// CountIncrement empty implementation.
func (n *Noop) CountIncrement(cached bool) {
}

// QueryIncrement empty implementation.
func (n *Noop) QueryIncrement(cached bool) {
}
