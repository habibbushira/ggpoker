package p2p

import (
	"sort"
	"strconv"
	"sync"
)

type PlayersList struct {
	lock sync.RWMutex
	list []string
}

func NewPlayersList() *PlayersList {
	return &PlayersList{
		list: []string{},
	}
}

func (p *PlayersList) List() []string {
	p.lock.RLock()
	p.lock.RUnlock()

	return p.list
}

func (p *PlayersList) getIndex(addr string) int {
	p.lock.Lock()
	defer p.lock.Unlock()

	for i := 0; i < len(p.list); i++ {
		if addr == p.list[i] {
			return i
		}
	}
	return -1
}

func (p *PlayersList) add(addr string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.list = append(p.list, addr)

	sort.Sort(p)

}

func (p *PlayersList) get(index int) string {
	p.lock.RLock()
	p.lock.RUnlock()

	if len(p.list)-1 < index {
		panic("index out of bounds")
	}

	return p.list[index]

}

func (p *PlayersList) Len() int {
	return len(p.list)
}

func (p *PlayersList) Swap(i, j int) {
	p.list[i], p.list[j] = p.list[j], p.list[i]
}

func (p *PlayersList) Less(i, j int) bool {
	portI, _ := strconv.Atoi(p.list[i][1:])
	portJ, _ := strconv.Atoi(p.list[j][1:])

	return portI < portJ
}
