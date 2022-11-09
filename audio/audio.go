package audio

import (
	soundconfig "github.com/iskrapw/sound/config"
)

type Sample float64

const Silence Sample = 0

type Audio interface {
	Initialize(config soundconfig.Config) error
	Dispose()
	Open() error
	Close()
	InputCallback(callback func(samples []Sample))
	OutputCallback(callback func(samples []Sample))
}
