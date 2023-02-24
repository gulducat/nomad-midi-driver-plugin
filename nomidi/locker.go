package nomidi

import (
	"sync"
)

type Locker struct {
	locks map[string]*sync.Mutex
	mut   *sync.Mutex
}

func NewLocker() *Locker {
	return &Locker{
		locks: make(map[string]*sync.Mutex),
		mut:   &sync.Mutex{},
	}
}

func (l *Locker) Get(name string) *sync.Mutex {
	l.mut.Lock()
	defer l.mut.Unlock()
	if lock, ok := l.locks[name]; ok {
		return lock
	} else {
		lock := &sync.Mutex{}
		l.locks[name] = lock
		return lock
	}
}

func (l *Locker) Lock(name string) (unlock func()) {
	//log.Println("LOCKING", name)
	lock := l.Get(name)
	lock.Lock()
	return func() {
		//log.Println("UNlocking", name)
		lock.Unlock()
	}
}
