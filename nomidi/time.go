package nomidi

import (
	"context"
	"errors"
	"github.com/hashicorp/go-hclog"
	"log"
	"sync"
	"time"
)

// this in-memory metronome is more doable in a week than a proper distributed clock..

// global map of clocks; keys are "song names" to keep them n'sync
var clocks = new(sync.Map)

type Clock struct {
	Name    string
	ticker  *time.Ticker
	locker  *Locker
	ticking bool
	stop    chan struct{}
	subs    map[string]*Player
	logger  hclog.Logger
}

func NewClock(name string) *Clock {
	l := hclog.Default().Named("Clock " + name)
	l.SetLevel(hclog.Debug)

	bpm := 120 // TODO
	l.Debug("BPM IS HARD-CODED!", "bpm", bpm)

	dur := time.Duration(bpm)
	bar := 60 * time.Second / dur * 4 // this is one bar in 4/4 time
	c := &Clock{
		Name:   name,
		ticker: time.NewTicker(bar),
		locker: NewLocker(),
		stop:   make(chan struct{}),
		subs:   make(map[string]*Player),
		logger: l,
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
	defer c.locker.Lock("subs")()
	if len(c.subs) != 0 {
		return errors.New("clock has subscribers")
	}
	c.logger.Info("i'm being deleted, byeee")
	c.Stop()
	clocks.Delete(name)
	return nil
}

func (c *Clock) Subscribe(p *Player) {
	c.logger.Info("subscribing", "port", p.Cfg.PortName)
	defer c.locker.Lock("subs")()
	c.subs[p.Cfg.PortName] = p
	c.logger.Info("subscribers", "subs", c.subs)
}

func (c *Clock) Unsubscribe(p *Player) {
	c.logger.Info("unsubscribing", "port", p.Cfg.PortName)
	defer c.locker.Lock("subs")()
	delete(c.subs, p.Cfg.PortName)
	c.logger.Info("subscribers", "subs", c.subs)
}

func (c *Clock) Stop() {
	close(c.stop)
}

func (c *Clock) Tick(ctx context.Context) {
	// is already ticking?
	unlock := c.locker.Lock("ticking")
	if c.ticking {
		unlock()
		return
	}
	c.ticking = true
	unlock()

	// if a Player stops listening to ticks, we can't keep sending to them, so set a timeout on sends.
	// it needs to be short to not interrupt music too badly, but not so short that it happens to active Players.
	tickTimeout := time.Millisecond * 100

	for {
		select {
		case <-c.stop:
			c.logger.Info("stopped")
			return
		case <-ctx.Done():
			c.logger.Warn("DID YOU KNOW YOUR CONTEXT STOPPED ME?")
			return
		case <-c.ticker.C:
			c.logger.Info("tick")
			unlock = c.locker.Lock("subs")
			// here there be time dragons
			for _, p := range c.subs {
				//c.logger.Debug("trying to send to", "port", p.Cfg.PortName)
				select {
				case p.Tick <- struct{}{}:
					//c.logger.Debug("sent to", "port", p.Cfg.PortName)
				case <-time.After(tickTimeout):
					// TODO: unsubscribe?  and mark as error'd so nomad can reschedule?  something else is wrong here..
					// making unsubscribe happen after wait results in this happening only once.
					//log.Printf("timeout sending Tick to %s", p.Cfg.PortName)
					c.logger.Warn("timeout sending Tick", "port", p.Cfg.PortName)
				}
			}
			unlock()
		}
	}
}
