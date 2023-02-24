package nomidi

import (
	"context"
	"github.com/hashicorp/nomad/client/structs"
	"github.com/hashicorp/nomad/plugins/device"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/plugins/drivers"
)

// taskHandle should store all relevant runtime information
// such as process ID if this is a local task or other meta
// data if this driver deals with external APIs
type taskHandle struct {
	// stateLock syncs access to all fields below
	stateLock sync.RWMutex

	logger      hclog.Logger
	taskConfig  *drivers.TaskConfig
	procState   drivers.TaskState
	startedAt   time.Time
	completedAt time.Time
	exitResult  *drivers.ExitResult

	clock   *Clock
	player  *Player
	stopper func()
}

func (h *taskHandle) TaskStatus() *drivers.TaskStatus {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()

	return &drivers.TaskStatus{
		ID:          h.taskConfig.ID,
		Name:        h.taskConfig.Name,
		State:       h.procState,
		StartedAt:   h.startedAt,
		CompletedAt: h.completedAt,
		ExitResult:  h.exitResult,
		//DriverAttributes: map[string]string{
		//	"pid": strconv.Itoa(h.pid),
		//},
	}
}

func (h *taskHandle) IsRunning() bool {
	h.stateLock.RLock()
	defer h.stateLock.RUnlock()
	return h.procState == drivers.TaskStateRunning
}

func (h *taskHandle) run(ctx context.Context) {
	h.stateLock.Lock()
	if h.exitResult == nil {
		h.exitResult = &drivers.ExitResult{}
	}
	h.stateLock.Unlock()

	h.stateLock.Lock()
	defer h.stateLock.Unlock()
	err := h.player.Wait(ctx)

	h.completedAt = time.Now()
	h.procState = drivers.TaskStateExited
	if err != nil {
		h.exitResult.Err = err
		h.exitResult.ExitCode = 1
		h.procState = drivers.TaskStateUnknown
	}
}

func (h *taskHandle) stats(ctx context.Context, interval time.Duration) (ch chan *drivers.TaskResourceUsage) {
	ch = make(chan *drivers.TaskResourceUsage)
	go func() {
		defer close(ch)
		timer := time.NewTimer(0)
		for {
			// all this for bogus zero stats...
			st := &drivers.TaskResourceUsage{
				ResourceUsage: &structs.ResourceUsage{
					CpuStats:    &structs.CpuStats{},
					MemoryStats: &structs.MemoryStats{},
					DeviceStats: []*device.DeviceGroupStats{},
				},
				Timestamp: time.Now().Unix(),
			}
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				timer.Reset(interval)
			case ch <- st:
			}
		}
	}()
	return ch
}
