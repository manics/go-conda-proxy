// String set
package repodata

var exists = struct{}{}

// https://www.davidkaya.com/p/sets-in-golang
type Set struct {
	items map[string]struct{}
}

func NewSet(items *[]string) *Set {
	s := &Set{}
	s.items = make(map[string]struct{})
	if items != nil {
		for _, i := range *items {
			s.Add(i)
		}
	}
	return s
}

func (s *Set) Add(value string) {
	s.items[value] = exists
}

func (s *Set) Pop() (string, bool) {
	for item := range s.items {
		s.Remove(item)
		return item, true
	}
	return "", false
}

func (s *Set) Remove(value string) {
	delete(s.items, value)
}

func (s *Set) Len() int {
	return len(s.items)
}

func (s *Set) Contains(value string) bool {
	_, c := s.items[value]
	return c
}

func (s *Set) Items() *[]string {
	items := []string{}
	for i := range s.items {
		items = append(items, i)
	}
	return &items
}
