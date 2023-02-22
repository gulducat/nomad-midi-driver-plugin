package nomidi

import (
	"context"
	"errors"
	"github.com/hashicorp/go-hclog"
	midi "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"log"
)

func NewPlayer(logger hclog.Logger, cfg TaskConfig) *Player {
	p := &Player{
		Cfg:   cfg,
		Tick:  make(chan struct{}),
		Done:  make(chan struct{}),
		errCh: make(chan error, 1),
		err:   errors.New("midi not done yet"),
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
	var err error
	select {
	case <-p.Done:
		err = p.Err()
	case <-ctx.Done():
		err = ctx.Err()
	}
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
	defer close(p.Done)

	port := p.Cfg.PortName
	file := p.Cfg.MidiFile
	bars := p.Cfg.Bars

	out, err := midi.FindOutPort(port)
	if err != nil {
		p.errCh <- err
		return
	}
	p.log.Info("found port", "out", out)
	defer func() {
		if err := out.Close(); err != nil {
			log.Printf("err closing out port %s: %s", out, err)
		}
	}()

	errCh := make(chan error, 1)
	bar := 1
	for {
		select {
		case <-ctx.Done():
			p.errCh <- ctx.Err() // is this an error for me, really?
			// this one is for operators, not job authors, so log instead of logger
			log.Printf("ctx done, so i (%s) am done too: %s", port, ctx.Err())
			return
		case e := <-errCh:
			p.errCh <- e
			log.Printf("error in player: %s", e)
			return
		case <-p.Tick:
			// clock says go ahead, once per bar.
		}

		// but we only play if lined up on the right bar count
		if bar > 1 {
			log.Printf("bar %d skip: %s", bar, port)
			bar--
			continue
		}
		bar = bars

		// for easier inspection in nomad agent logs for now
		log.Printf("bar %d play: %s", bar, port)
		// this goes to task logs
		//p.log.Info("playing")

		//err = smf.ReadTracks(file).Do(
		//	func(te smf.TrackEvent) {
		//		p.log.Info("te",
		//			"track", te.TrackNo,
		//			"msg", te.Message.String(),
		//		)
		//	},
		//).Play(out)

		// this blocks so without a goroutine would produce variable duration between tick reads.
		// backgrounding this allows the clock to continue ticking appropriately.
		// TODO: but now this not-blocking means the program can exit without a MIDI NOTE OFF command...
		go func() {
			if e := smf.ReadTracks(file).Play(out); e != nil {
				errCh <- e
			}
		}()
	}
}
