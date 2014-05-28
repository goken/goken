package instrumentor

type stackEntry struct {
	next *stackEntry
	value interface{}
}

type stack struct {
	top *stackEntry
	count int
}

func (s *stack) Push(v interface{}) {
	var e stackEntry
	e.value = v
	e.next = s.top
	s.top = &e
	s.count++
}

func (s *stack) Pop() (v interface{}) {
	v = s.Peek()
	if v != nil {
		s.top = s.top.next
		s.count--
	}
	return
}

func (s *stack) Peek() (v interface{}) {
	if s.top == nil {
		return nil
	}
	v = s.top.value
	return
}

func (s *stack) Count() int {
	return s.count
}

type Stack interface {
	Push(interface{})
	Pop() interface{}
	Peek() interface{}
	Count() int
}

func NewStack() Stack {
	return &stack{}
}
