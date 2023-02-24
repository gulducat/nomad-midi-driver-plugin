package maestro

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/nomad/client/lib/fifo"
	"os"
	"path/filepath"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/drivers/shared/eventer"
	"github.com/hashicorp/nomad/plugins/base"
	"github.com/hashicorp/nomad/plugins/drivers"
	"github.com/hashicorp/nomad/plugins/shared/hclspec"
	"github.com/hashicorp/nomad/plugins/shared/structs"
)

const (
	// pluginName is the name of the plugin
	// this is used for logging and (along with the version) for uniquely
	// identifying plugin binaries fingerprinted by the client
	pluginName = "midi-portmidi"

	// pluginVersion allows the client to identify and use newer versions of
	// an installed plugin
	pluginVersion = "v0.1.0"

	// fingerprintPeriod is the interval at which the plugin will send
	// fingerprint responses
	fingerprintPeriod = 30 * time.Second

	// taskHandleVersion is the version of task handle which this plugin sets
	// and understands how to decode
	// this is used to allow modification and migration of the task schema
	// used by the plugin
	taskHandleVersion = 1
)

var (
	// binPath is the full path to the binary that nomad runs
	binPath = os.Args[0]

	pluginDir = filepath.Dir(binPath) // laaazy

	// pluginInfo describes the plugin
	pluginInfo = &base.PluginInfoResponse{
		Type:              base.PluginTypeDriver,
		PluginApiVersions: []string{drivers.ApiVersion010},
		PluginVersion:     pluginVersion,
		Name:              pluginName,
	}

	// configSpec is the specification of the plugin's configuration
	// this is used to validate the configuration specified for the plugin
	// on the client.
	// this is not global, but can be specified on a per-client basis.
	configSpec = hclspec.NewObject(map[string]*hclspec.Spec{})

	// taskConfigSpec is the specification of the plugin's configuration for
	// a task
	// this is used to validated the configuration specified for the plugin
	// when a job is submitted.
	taskConfigSpec = hclspec.NewObject(map[string]*hclspec.Spec{
		"song":      hclspec.NewAttr("song", "string", true),
		"midi_file": hclspec.NewAttr("midi_file", "string", true),
		"port_name": hclspec.NewAttr("port_name", "string", true),
		"bars":      hclspec.NewAttr("bars", "number", true),
	})

	// capabilities indicates what optional features this driver supports
	// this should be set according to the target run time.
	capabilities = &drivers.Capabilities{
		// https://godoc.org/github.com/hashicorp/nomad/plugins/drivers#Capabilities
		SendSignals: false,
		Exec:        false,
	}
)

// Config contains configuration information for the plugin
type Config struct {
}

// TaskConfig contains configuration information for a task that runs with
// this plugin
type TaskConfig struct {
	Song     string `codec:"song"`
	MidiFile string `codec:"midi_file"`
	PortName string `codec:"port_name"`
	Bars     int    `codec:"bars"`
}

// TaskState is the runtime state which is encoded in the handle returned to
// Nomad client.
// This information is needed to rebuild the task state and handler during
// recovery.
type TaskState struct {
	ReattachConfig *structs.ReattachConfig
	TaskConfig     *drivers.TaskConfig
	StartedAt      time.Time

	// TODO: add any extra important values that must be persisted in order
	// to restore a task.
}

// MIDIDriverPlugin tasks will play a midi file through a midi port.
type MIDIDriverPlugin struct {
	// eventer is used to handle multiplexing of TaskEvents calls such that an
	// event can be broadcast to all callers
	eventer *eventer.Eventer

	// config is the plugin configuration set by the SetConfig RPC
	config *Config

	// nomadConfig is the client config from Nomad
	nomadConfig *base.ClientDriverConfig

	// tasks is the in memory datastore mapping taskIDs to driver handles
	tasks *taskStore

	// ctx is the context for the driver. It is passed to other subsystems to
	// coordinate shutdown
	ctx context.Context

	// signalShutdown is called when the driver is shutting down and cancels
	// the ctx passed to any subsystems
	signalShutdown context.CancelFunc

	// logger will log to the Nomad agent
	logger hclog.Logger
}

// NewPlugin returns a new example driver plugin
func NewPlugin(logger hclog.Logger) drivers.DriverPlugin {
	ctx, cancel := context.WithCancel(context.Background())
	logger = logger.Named(pluginName)

	return &MIDIDriverPlugin{
		eventer:        eventer.NewEventer(ctx, logger),
		config:         &Config{},
		tasks:          newTaskStore(),
		ctx:            ctx,
		signalShutdown: cancel,
		logger:         logger,
	}
}

// PluginInfo returns information describing the plugin.
func (d *MIDIDriverPlugin) PluginInfo() (*base.PluginInfoResponse, error) {
	return pluginInfo, nil
}

// ConfigSchema returns the plugin configuration schema.
func (d *MIDIDriverPlugin) ConfigSchema() (*hclspec.Spec, error) {
	return configSpec, nil
}

// SetConfig is called by the client to pass the configuration for the plugin.
func (d *MIDIDriverPlugin) SetConfig(cfg *base.Config) error {
	var config Config
	if len(cfg.PluginConfig) != 0 {
		if err := base.MsgPackDecode(cfg.PluginConfig, &config); err != nil {
			return err
		}
	}

	// Save the configuration to the plugin
	d.config = &config

	// TODO: parse and validated any configuration value if necessary.

	// Save the Nomad agent configuration
	if cfg.AgentConfig != nil {
		d.nomadConfig = cfg.AgentConfig.Driver
	}

	return nil
}

// TaskConfigSchema returns the HCL schema for the configuration of a task.
func (d *MIDIDriverPlugin) TaskConfigSchema() (*hclspec.Spec, error) {
	return taskConfigSpec, nil
}

// Capabilities returns the features supported by the driver.
func (d *MIDIDriverPlugin) Capabilities() (*drivers.Capabilities, error) {
	return capabilities, nil
}

// Fingerprint returns a channel that will be used to send health information
// and other driver specific node attributes.
func (d *MIDIDriverPlugin) Fingerprint(ctx context.Context) (<-chan *drivers.Fingerprint, error) {
	ch := make(chan *drivers.Fingerprint)
	go d.handleFingerprint(ctx, ch)
	return ch, nil
}

// handleFingerprint manages the channel and the flow of fingerprint data.
func (d *MIDIDriverPlugin) handleFingerprint(ctx context.Context, ch chan<- *drivers.Fingerprint) {
	defer close(ch)

	// Nomad expects the initial fingerprint to be sent immediately
	ticker := time.NewTimer(0)
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			// after the initial fingerprint we can set the proper fingerprint
			// period
			ticker.Reset(fingerprintPeriod)
			ch <- d.buildFingerprint()
		}
	}
}

// buildFingerprint returns the driver's fingerprint data
func (d *MIDIDriverPlugin) buildFingerprint() *drivers.Fingerprint {
	fp := &drivers.Fingerprint{
		Attributes:        map[string]*structs.Attribute{},
		Health:            drivers.HealthStateHealthy,
		HealthDescription: drivers.DriverHealthy,
	}

	// TODO: implement fingerprinting logic to populate health and driver
	// attributes.
	//
	// Fingerprinting is used by the plugin to relay two important information
	// to Nomad: health state and node attributes.
	//
	// If the plugin reports to be unhealthy, or doesn't send any fingerprint
	// data in the expected interval of time, Nomad will restart it.
	//
	// Node attributes can be used to report any relevant information about
	// the node in which the plugin is running (specific library availability,
	// installed versions of a software etc.). These attributes can then be
	// used by an operator to set job constrains.
	//
	// In the example below we check if the shell specified by the user exists
	// in the node.
	//shell := d.config.LockFile

	//cmd := exec.Command("which", shell)
	//if err := cmd.Run(); err != nil {
	//	return &drivers.Fingerprint{
	//		Health:            drivers.HealthStateUndetected,
	//		HealthDescription: fmt.Sprintf("shell %s not found", shell),
	//	}
	//}

	//// We also set the shell and its version as attributes
	//cmd = exec.Command(shell, "--version")
	//if out, err := cmd.Output(); err != nil {
	//	d.logger.Warn("failed to find shell version: %v", err)
	//} else {
	//	re := regexp.MustCompile("[0-9]\\.[0-9]\\.[0-9]")
	//	version := re.FindString(string(out))

	//	fp.Attributes["driver.hello.shell_version"] = structs.NewStringAttribute(version)
	//	fp.Attributes["driver.hello.lock_file"] = structs.NewStringAttribute(shell)
	//}

	return fp
}

// StartTask returns a task handle and a driver network if necessary.
func (d *MIDIDriverPlugin) StartTask(cfg *drivers.TaskConfig) (*drivers.TaskHandle, *drivers.DriverNetwork, error) {
	if _, ok := d.tasks.Get(cfg.ID); ok {
		return nil, nil, fmt.Errorf("task with ID %q already started", cfg.ID)
	}

	//d.logger.Error("HI TASK CONFIG", "cfg", fmt.Sprintf("%#v", cfg))

	var driverConfig TaskConfig
	if err := cfg.DecodeDriverConfig(&driverConfig); err != nil {
		return nil, nil, fmt.Errorf("failed to decode driver config: %v", err)
	}

	d.logger.Info("starting task", "driver_cfg", hclog.Fmt("%+v", driverConfig))
	handle := drivers.NewTaskHandle(taskHandleVersion)
	handle.Config = cfg

	ctx, stopper := context.WithCancel(d.ctx)

	stdout, err := fifo.OpenWriter(cfg.StdoutPath)
	if err != nil {
		return nil, nil, fmt.Errorf("fifo.OpenWriter err: %w", err)
	}
	opts := &hclog.LoggerOptions{
		Output: stdout,
		Level:  hclog.Debug,
		TimeFn: time.Now,
	}
	logger := hclog.New(opts)

	clock := GetClock(driverConfig.Song)
	player := NewPlayer(logger, driverConfig)
	clock.Subscribe(player)

	h := &taskHandle{
		taskConfig: cfg,
		procState:  drivers.TaskStateRunning,
		startedAt:  time.Now().Round(time.Millisecond),
		logger:     d.logger,
		// my stuff
		clock:   clock,
		player:  player,
		stopper: stopper,
	}

	driverState := TaskState{
		TaskConfig: cfg,
		StartedAt:  h.startedAt,
	}

	if err := handle.SetDriverState(&driverState); err != nil {
		return nil, nil, fmt.Errorf("failed to set driver state: %v", err)
	}

	d.tasks.Set(cfg.ID, h)

	// !! the clock ticking *needs* not to be tied to any single task, so it gets context.Background()
	// it gets cleaned up in DestroyTask() when there are no subscribers to it.
	go clock.Tick(context.Background())

	go player.Play(ctx)
	go h.run(ctx)
	return handle, nil, nil
}

// RecoverTask recreates the in-memory state of a task from a TaskHandle.
func (d *MIDIDriverPlugin) RecoverTask(handle *drivers.TaskHandle) error {
	if handle == nil {
		return errors.New("error: handle cannot be nil")
	}

	if _, ok := d.tasks.Get(handle.Config.ID); ok {
		return nil
	}

	// TODO(db): replicate this happening, what can be done?
	return fmt.Errorf("unable to recover midi task like this...? id: %s", handle.Config.ID)

	var taskState TaskState
	if err := handle.GetDriverState(&taskState); err != nil {
		return fmt.Errorf("failed to decode task state from handle: %v", err)
	}

	var driverConfig TaskConfig
	if err := taskState.TaskConfig.DecodeDriverConfig(&driverConfig); err != nil {
		return fmt.Errorf("failed to decode driver config: %v", err)
	}

	h := &taskHandle{
		taskConfig: taskState.TaskConfig,
		procState:  drivers.TaskStateRunning,
		startedAt:  taskState.StartedAt,
		exitResult: &drivers.ExitResult{},
	}

	d.tasks.Set(taskState.TaskConfig.ID, h)

	go h.run(d.ctx)
	return nil
}

// WaitTask returns a channel used to notify Nomad when a task exits.
func (d *MIDIDriverPlugin) WaitTask(ctx context.Context, taskID string) (<-chan *drivers.ExitResult, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	ch := make(chan *drivers.ExitResult)
	go d.handleWait(ctx, handle, ch)
	return ch, nil
}

func (d *MIDIDriverPlugin) handleWait(ctx context.Context, handle *taskHandle, ch chan *drivers.ExitResult) {
	defer close(ch)
	var result = &drivers.ExitResult{}

	err := handle.player.Wait(ctx)
	if err != nil {
		result.Err = fmt.Errorf("handleWait: error waiting on player: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case ch <- result:
		}
	}
}

// StopTask stops a running task with the given signal and within the timeout window.
func (d *MIDIDriverPlugin) StopTask(taskID string, timeout time.Duration, signal string) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	handle.logger.Info("stopping task",
		"id", taskID,
		"timeout", timeout.String(),
		"signal", signal)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	handle.clock.Unsubscribe(handle.player)
	handle.stopper()
	handle.player.Wait(ctx)

	return nil
}

// DestroyTask cleans up and removes a task that has terminated.
func (d *MIDIDriverPlugin) DestroyTask(taskID string, force bool) error {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return drivers.ErrTaskNotFound
	}

	if handle.IsRunning() && !force {
		return errors.New("cannot destroy running task")
	}

	select {
	case <-handle.player.Done:
		handle.logger.Info("deleting", "id", taskID)
		d.tasks.Delete(taskID)
		if err := DeleteClock(handle.clock.Name); err != nil {
			handle.logger.Info("did not delete clock", "err", err)
		}
	default:
		// do i ever hit this?
		return errors.New("player context not done")
	}

	return nil
}

// InspectTask returns detailed status information for the referenced taskID.
func (d *MIDIDriverPlugin) InspectTask(taskID string) (*drivers.TaskStatus, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}

	return handle.TaskStatus(), nil
}

// TaskStats returns a channel which the driver should send stats to at the given interval.
func (d *MIDIDriverPlugin) TaskStats(ctx context.Context, taskID string, interval time.Duration) (<-chan *drivers.TaskResourceUsage, error) {
	handle, ok := d.tasks.Get(taskID)
	if !ok {
		return nil, drivers.ErrTaskNotFound
	}
	return handle.stats(ctx, interval), nil
}

// TaskEvents returns a channel that the plugin can use to emit task related events.
func (d *MIDIDriverPlugin) TaskEvents(ctx context.Context) (<-chan *drivers.TaskEvent, error) {
	return d.eventer.TaskEvents(ctx)
}

// SignalTask forwards a signal to a task.
// This is an optional capability.
func (d *MIDIDriverPlugin) SignalTask(taskID string, signal string) error {
	return nil
}

// ExecTask returns the result of executing the given command inside a task.
// This is an optional capability.
func (d *MIDIDriverPlugin) ExecTask(taskID string, cmd []string, timeout time.Duration) (*drivers.ExecTaskResult, error) {
	// TODO: implement driver specific logic to execute commands in a task.
	return nil, errors.New("This driver does not support exec")
}
