package maestro

import (
	"context"
	"errors"
	"github.com/hashicorp/go-hclog"
	midi "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"path"
	"sync"
	"time"
)

func NewPlayer(logger hclog.Logger, cfg TaskConfig) *Player {
	logger = hclog.Default().Named("Player " + cfg.PortName)
	logger.SetLevel(hclog.Debug)
	p := &Player{
		Cfg:   cfg,
		Tick:  make(chan struct{}, 1),
		Done:  make(chan struct{}),
		errCh: make(chan error, 1),
		log:   logger,
	}
	return p
}

type Player struct {
	Cfg   TaskConfig
	Tick  chan struct{}
	Done  chan struct{}
	errCh chan error
	err   error
	log   hclog.Logger
}

func (p *Player) Wait(ctx context.Context) error {
	p.log.Debug("start waiting")
	var err error
	select {
	case <-p.Done:
		err = p.Err()
	case <-ctx.Done():
		err = ctx.Err()
	}
	p.log.Debug("done waiting", "err", err)
	return err
}

func (p *Player) Err() error {
	select {
	case e := <-p.errCh:
		p.err = e
	default:
	}
	return p.err
}

func (p *Player) Play(ctx context.Context) {
	port := p.Cfg.PortName
	file := path.Join(pluginDir, p.Cfg.MidiFile)
	bars := p.Cfg.Bars

	// connect first
	out, err := midi.FindOutPort(port)
	if err != nil {
		p.errCh <- err
		close(p.Done)
		return
	}
	p.log.Info("found out port", "port", out)
	// close last
	defer func() {
		if err := out.Close(); err != nil {
			p.log.Error("error closing out port", "port", port, "err", err)
		}
		p.log.Info("port closed", "port", port)
	}()

	// silly dance to ensure the midi file finishes before we call the player done
	var wg sync.WaitGroup
	defer func() {
		wg.Wait()
		close(p.Done)
	}()

	errCh := make(chan error, 1)
	bar := 1
	for {
		select {
		case <-ctx.Done():
			p.log.Info("ctx done, so i am done too", "port", port)
			return
		case e := <-errCh:
			p.errCh <- e
			p.log.Error("player error", "err", e)
			return
		case <-time.After(time.Second * 10): // 10 is ok? would be down to like 30 bpm
			p.errCh <- errors.New("timeout after 5 seconds waiting for Tick")
			return
		case <-p.Tick:
			// clock says go ahead, once per bar.
		}

		// but we only play if lined up on the right bar count
		if bar > 1 {
			p.log.Debug("skipping", "port", port, "bar", bar)
			bar--
			continue
		}
		bar = bars

		p.log.Info("playing", "port", port, "bar", bar)

		// ReadTracks() blocks so without a goroutine would produce variable duration between tick reads.
		// backgrounding this allows the clock to continue ticking appropriately.
		wg.Add(1)
		go func() {
			defer wg.Done()
			err = smf.ReadTracks(file).Play(out)
			//err = smf.ReadTracks(file).Do(
			//	func(te smf.TrackEvent) {
			//		log.Printf("port %s: %s; %#v", port, te.Message.String(), te)
			//	},
			//).Play(out)
			if err != nil {
				errCh <- err
			}
		}()
	}
}
