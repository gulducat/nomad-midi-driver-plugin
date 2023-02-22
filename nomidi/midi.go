package nomidi

import (
	"context"
	"errors"
	"github.com/hashicorp/go-hclog"
	midi "gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"log"
)

func NewPlayer(logger hclog.Logger) *Player {
	return &Player{
		Done:  make(chan struct{}),
		errCh: make(chan error, 1),
		err:   errors.New("midi not done yet"),
		log:   logger,
	}
}

type Player struct {
	Done  chan struct{}
	errCh chan error
	err   error
	log   hclog.Logger
}

func (p *Player) Wait(ctx context.Context) error {
	select {
	case <-p.Done:
		return p.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (p *Player) Err() error {
	select {
	case e := <-p.errCh:
		p.err = e
	default:
	}
	return p.err
}

func (p *Player) Play(ctx context.Context, port, file string) {
	defer close(p.Done)

	out, err := midi.FindOutPort(port)
	if err != nil {
		p.errCh <- err
		return
	}
	p.log.Info("found port", "out", out)
	for {
		select {
		case <-ctx.Done():
			p.errCh <- ctx.Err() // is this an error for me, really?
			// this one is for operators, not job authors, so log instead of logger
			log.Printf("ctx done, so i'm done too: %s", ctx.Err())
			return
		default:
		}

		err = smf.ReadTracks(file).Do(
			func(te smf.TrackEvent) {
				p.log.Info("te",
					"track", te.TrackNo,
					"msg", te.Message.String(),
				)
			},
		).Play(out)
		if err != nil {
			p.errCh <- err
			return
		}
	}
}
