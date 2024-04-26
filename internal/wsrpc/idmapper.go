package wsrpc

import "sync"

type IdMapper struct {
	fw map[string]uint
	bw map[uint]string
	mu sync.RWMutex
}

func NewIdMapper() *IdMapper {
	return &IdMapper{
		fw: make(map[string]uint),
		bw: make(map[uint]string),
	}
}

func (m *IdMapper) Add(inner string, outer uint) {
	m.mu.Lock()
	m.fw[inner] = outer
	m.bw[outer] = inner
	m.mu.Unlock()
}

func (m *IdMapper) RemoveByOuter(outer uint) {
	m.mu.Lock()
	inner := m.bw[outer]
	delete(m.bw, outer)
	delete(m.fw, inner)
	m.mu.Unlock()
}

func (m *IdMapper) ToInner(outer uint) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	inner, e := m.bw[outer]
	return inner, e
}

func (m *IdMapper) ToOuter(inner string) (uint, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	outer, e := m.fw[inner]
	return outer, e
}
