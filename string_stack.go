package urknall

import "strings"

type rlStack struct {
	data []string
	size int
}

func (s *rlStack) Push(val string) {
	switch {
	case s.size < len(s.data):
		s.data[s.size] = val
	default:
		s.data = append(s.data, val)
	}
	s.size += 1
}

func (s *rlStack) Pop() string {
	if s.size > 0 {
		s.size -= 1
		return s.data[s.size]
	}
	return ""
}

func (s *rlStack) String() string {
	return strings.Join(s.data[:s.size], ".")
}
