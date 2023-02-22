package nomidi

import (
	"context"
	"errors"
	"github.com/hashicorp/go-hclog"
	midi "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"log"
)

// TODO: just pass cfg in here.
func NewPlayer(logger hclog.Logger, port, file string) *Player {
	p := &Player{
		Port:  port,
		File:  file,
		Tick:  make(chan struct{}, 1),
		Done:  make(chan struct{}),
		errCh: make(chan error, 1),
		err:   errors.New("midi not done yet"),
		log:   logger,
	}
	return p
}

type Player struct {
	Port  string
	File  string
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

	out, err := midi.FindOutPort(p.Port)
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

	for {
		select {
		case <-ctx.Done():
			p.errCh <- ctx.Err() // is this an error for me, really?
			// this one is for operators, not job authors, so log instead of logger
			log.Printf("ctx done, so i (%s) am done too: %s", p.Port, ctx.Err())
			return
		case <-p.Tick:
			// clock says go ahead.
		}

		// for easier inspection in nomad agent logs for now
		log.Println("playing:", p.Port)
		// this goes to task logs
		//p.log.Info("playing")

		//err = smf.ReadTracks(p.File).Do(
		//	func(te smf.TrackEvent) {
		//		p.log.Info("te",
		//			"track", te.TrackNo,
		//			"msg", te.Message.String(),
		//		)
		//	},
		//).Play(out)
		err = smf.ReadTracks(p.File).Play(out)
		if err != nil {
			p.errCh <- err
			return
		}
	}
}
