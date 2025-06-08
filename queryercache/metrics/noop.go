package metrics

type Noop struct {
}

func NewNoop() *Noop {
	return &Noop{}
}

func (n *Noop) CountIncrement(cached bool) {
}

func (n *Noop) QueryIncrement(cached bool) {
}
