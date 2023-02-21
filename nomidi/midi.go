package nomidi

import (
	"fmt"
	midi "gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregisters driver
	"gitlab.com/gomidi/midi/v2/smf"
	"log"
)

func Play(port, file string) error {
	defer midi.CloseDriver()

	/* this already gets caught by ReadTracks...
	// and methinks we already need to CloseDriver() ...?
	if _, err := os.Stat(file); errors.Is(err, os.ErrNotExist) {
		return err
	}
	*/

	out, err := midi.FindOutPort(port)
	if err != nil {
		return err
	}
	log.Println("out:", out)
	for {
		err = smf.ReadTracks(file).Do(
			func(te smf.TrackEvent) {
				if te.Message.IsMeta() {
					fmt.Printf("[%v] @%vms %s\n", te.TrackNo, te.AbsMicroSeconds/1000, te.Message.String())
					/*
						var t string
						if mm.Text(&t) {
							//fmt.Printf("[%v] %s %s (%s): %q\n", te.TrackNo, msg.Type().Kind(), msg.String(), msg.Type(), t)
							fmt.Printf("[%v] %s: %q\n", te.TrackNo, te.Type, t)
							//fmt.Printf("[%v] %s %s (%s): %q\n", te.TrackNo, mm.Type().Kind(), mm.String(), mm.Type(), t)
						}
						var bpm float64
						if mm.Tempo(&bpm) {
							fmt.Printf("[%v] %s: %v\n", te.TrackNo, te.Type, math.Round(bpm))
						}
					*/
				} else {
					fmt.Printf("[%v] %s\n", te.TrackNo, te.Message)
				}
			},
		).Play(out)
		if err != nil {
			return err
		}
	}
	return nil
}

// this junk is already done in state.go -> taskStore
/*
import "sync"

// global state
var tasks *sync.Map = new(sync.Map)

type MIDI struct {
	ID string
}

func SetTask(m *MIDI) {
	tasks.Store(m.ID, m)
}

func GetTask(id string) *MIDI {
	t, ok := tasks.Load(id)
	if !ok {
		return new(MIDI)
	}
	return t.(*MIDI)
}

func DeleteTask(id string) {
	tasks.Delete(id)
}
*/
