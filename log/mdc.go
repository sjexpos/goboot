package log

import "github.com/sjexpos/goboot/concurrent"

var MDC mdc = mdc{}

type mdc struct {
	resource concurrent.GoRoutineLocal[map[string]string]
}

func (m *mdc) allValues() *map[string]string {
	values := m.resource.Get()
	if values == nil {
		v := make(map[string]string, 0)
		values = &v
		m.resource.Set(values)
	}
	return values
}

func (m *mdc) Set(key string, value string) {
	values := m.allValues()
	(*values)[key] = value
}

func (m *mdc) Get(key string) string {
	values := m.allValues()
	return (*values)[key]
}

func (m *mdc) Clean(key string) {
	values := m.allValues()
	delete((*values), key)
}
