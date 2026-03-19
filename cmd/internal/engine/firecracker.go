package engine

import (
	"context"
	"os"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	models "github.com/firecracker-microvm/firecracker-go-sdk/client/models"
	"github.com/sirupsen/logrus"
)

func init() {
	logFile, err := os.OpenFile("engine.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err == nil {
		logrus.SetOutput(logFile)
	}
}

type VM struct {
	machine *firecracker.Machine
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewVM(socketPath, kernelPath, rootfsPath string) (*VM, error) {
	ctx, cancel := context.WithCancel(context.Background())

	os.Remove(socketPath)
	os.Remove("/tmp/v.sock")

	cfg := firecracker.Config{
		SocketPath:      socketPath,
		KernelImagePath: kernelPath,
		KernelArgs:      "console=ttyS0 reboot=k panic=1 pci=off root=/dev/vda rw init=/sbin/init",
		Drives: []models.Drive{{
			DriveID: firecracker.String("1"), PathOnHost: firecracker.String(rootfsPath),
			IsRootDevice: firecracker.Bool(true), IsReadOnly: firecracker.Bool(false),
		}},
		VsockDevices: []firecracker.VsockDevice{{
			Path: "/tmp/v.sock",
			CID:  3,
		}},
		NetworkInterfaces: []firecracker.NetworkInterface{{
			StaticConfiguration: &firecracker.StaticNetworkConfiguration{
				MacAddress:  "AA:FC:00:00:00:01",
				HostDevName: "tap0",
			},
		}},
		MachineCfg: models.MachineConfiguration{
			VcpuCount:  firecracker.Int64(2),
			MemSizeMib: firecracker.Int64(1024),
		},
	}

	bootLog, err := os.OpenFile("firecracker_boot.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		bootLog, _ = os.OpenFile(os.DevNull, os.O_WRONLY, 0666)
	}

	cmd := firecracker.VMCommandBuilder{}.
		WithSocketPath(socketPath).
		WithStdout(bootLog).
		WithStderr(bootLog).
		Build(ctx)

	logger := logrus.NewEntry(logrus.StandardLogger())

	machine, err := firecracker.NewMachine(ctx, cfg, firecracker.WithProcessRunner(cmd), firecracker.WithLogger(logger))
	if err != nil {
		cancel()
		return nil, err
	}

	return &VM{machine: machine, ctx: ctx, cancel: cancel}, nil
}

func (v *VM) Start() error { return v.machine.Start(v.ctx) }
func (v *VM) Stop() error  { defer v.cancel(); return v.machine.StopVMM() }
