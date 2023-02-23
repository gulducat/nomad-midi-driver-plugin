package nomidi

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

// this in-memory metronome is more doable in a week than a proper distributed clock..

// keys are "song names" to keep them n'sync
var clocks = new(sync.Map)

type Clock struct {
	Name    string
	ticker  *time.Ticker
	lock    sync.Mutex
	ticking bool
	subs    map[string]chan struct{}
}

func NewClock(name string) *Clock {
	bpm := 120 // TODO
	log.Printf("BPM IS ALWAYS %d!", bpm)
	dur := time.Duration(bpm)
	bar := 60 * time.Second / dur * 4 // this is one bar in 4/4 time
	c := &Clock{
		Name:   name,
		ticker: time.NewTicker(bar),
		subs:   make(map[string]chan struct{}),
	}
	clocks.Store(name, c)
	return c
}

func GetClock(name string) *Clock {
	c, ok := clocks.Load(name)
	if ok {
		log.Println("found clock:", name)
		return c.(*Clock)
	}
	log.Println("new clock:", name)
	return NewClock(name)
}

func DeleteClock(name string) error {
	c := GetClock(name)
	c.lock.Lock()
	defer c.lock.Unlock()
	if len(c.subs) != 0 {
		return errors.New("clock has subscribers")
	}
	clocks.Delete(name)
	return nil
}

func (c *Clock) Subscribe(p *Player) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.subs[p.Cfg.PortName] = p.Tick
}

func (c *Clock) Unsubscribe(p *Player) {
	c.lock.Lock()
	defer c.lock.Unlock()
	delete(c.subs, p.Cfg.PortName)
}

func (c *Clock) Tick(ctx context.Context) {
	// is already ticking?
	c.lock.Lock()
	if c.ticking {
		c.lock.Unlock()
		return
	}
	c.ticking = true
	c.lock.Unlock()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.ticker.C:
			log.Println("tick:", c.Name)
			c.lock.Lock()
			// here there be time dragons
			for _, ch := range c.subs {
				// this lil goroutine causes ~125 milliseconds of delay. whoa!
				//go func() {
				ch <- struct{}{}
				//}()
			}
			c.lock.Unlock()
		}
	}
}
