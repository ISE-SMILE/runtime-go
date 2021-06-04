package openwhisk

import (
	"bufio"
	"encoding/json"
	"os"
	"os/signal"
	"syscall"
)

type LifeCycleHookFlags struct {
	Ok        bool `json:"ok"`
	Pausing   bool `json:"pause,omitempty"`
	Finishing bool `json:"finish,omitempty"`
	Hinting   bool `json:"hint,omitempty"`
	Freshen   bool `json:"freshen,omitempty"`
}

type LifeCycleHooks interface {
	Pause(out *os.File)
	Stop(out *os.File)
	Hint(in map[string]string, out *os.File)
	Freshen(in map[string]string, out *os.File)
}

type BaseSignals struct{}

func (b BaseSignals) Pause(out *os.File) {}

func (b BaseSignals) Stop(out *os.File) {}

func (b BaseSignals) Hint(in map[string]string, out *os.File) {}

func (b BaseSignals) Freshen(in map[string]string, out *os.File) {}

func ActivateHooks(supportedHooks LifeCycleHookFlags, hooks LifeCycleHooks, input *bufio.Reader, output *os.File) []byte {
	if hooks != nil {
		signals := make([]os.Signal, 0)
		if supportedHooks.Pausing {
			signals = append(signals, syscall.SIGINT)
		}

		if supportedHooks.Finishing {
			signals = append(signals, syscall.SIGABRT)
		}

		if supportedHooks.Hinting {
			signals = append(signals, syscall.SIGUSR1)
		}

		if supportedHooks.Freshen {
			signals = append(signals, syscall.SIGUSR2)
		}

		capture := make(chan os.Signal, 2)
		signal.Notify(capture, signals...)

		go func() {
			for {
				sig := <-capture
				switch sig {
				case syscall.SIGINT:
					hooks.Pause(output)
				case syscall.SIGABRT:
					hooks.Stop(output)
					return
				case syscall.SIGUSR1:
					hooks.Hint(nil, output)
				case syscall.SIGUSR2:
					hooks.Freshen(nil, output)
				}
			}
		}()
	}

	// acknowledgement of started action
	buf, err := json.Marshal(supportedHooks)
	if err != nil {
		return []byte("{ \"ok\": false  }")
	}

	return buf
}
