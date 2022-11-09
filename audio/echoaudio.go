package audio

import (
	"math/rand"
	"sync"

	"github.com/iskrapw/dsp/filter/butterworth"
	soundconfig "github.com/iskrapw/sound/config"
	"github.com/iskrapw/utils/misc"
)

type EchoAudio struct {
	operate          bool
	sampleBuffer     []Sample
	sampleBufferLock sync.Mutex

	inputCallback    func([]Sample)
	inputBufferSize  int
	inputLoopTimeout float64

	outputCallback    func([]Sample)
	outputBufferSize  int
	outputLoopTimeout float64

	outputSampleRate float64
	noiseLevel       float64
	highCut          float64
	lowCut           float64
	bandPassFilter   butterworth.BandPass
}

func (a *EchoAudio) Initialize(config soundconfig.Config) error {
	a.sampleBuffer = make([]Sample, 0)
	a.operate = false

	a.inputBufferSize = config.Input.BufferSize
	a.inputLoopTimeout = float64(a.inputBufferSize) / float64(config.Input.SampleRate)

	a.outputBufferSize = config.Output.BufferSize
	a.outputLoopTimeout = float64(a.outputBufferSize) / float64(config.Output.SampleRate)

	a.outputSampleRate = float64(config.Output.SampleRate)
	a.noiseLevel = config.Echo.NoiseLevel
	a.highCut = config.Echo.HighCut
	a.lowCut = config.Echo.LowCut
	a.bandPassFilter.Setup(config.Echo.BPFOrder, a.outputSampleRate, a.lowCut, a.highCut)
	return nil
}

func (a *EchoAudio) Dispose() {

}

func (a *EchoAudio) Open() error {
	a.operate = true
	go a.inputCallbackLoop()
	go a.outputCallbackLoop()
	return nil
}

func (a *EchoAudio) Close() {
	a.operate = false
}

func (a *EchoAudio) InputCallback(c func([]Sample)) {
	a.inputCallback = func(samples []Sample) {
		c(samples)
	}
}

func (a *EchoAudio) OutputCallback(c func([]Sample)) {
	a.outputCallback = func(samples []Sample) {
		c(samples)
	}
}

func (a *EchoAudio) inputCallbackLoop() {
	inputCallbackSamplesBuffer := make([]Sample, a.inputBufferSize)
	zerosBuffer := make([]Sample, a.inputBufferSize)
	for a.operate {
		for len(a.sampleBuffer) >= a.inputBufferSize {
			limit := a.inputBufferSize
			if limit > len(a.sampleBuffer) {
				limit = len(a.sampleBuffer)
			}
			copy(inputCallbackSamplesBuffer[0:limit], a.sampleBuffer[0:limit])
			copy(inputCallbackSamplesBuffer[limit:a.inputBufferSize], zerosBuffer[limit:a.inputBufferSize])
			a.simulateChannel(inputCallbackSamplesBuffer)

			a.sampleBufferLock.Lock()
			a.sampleBuffer = a.sampleBuffer[limit:]
			a.sampleBufferLock.Unlock()

			a.inputCallback(inputCallbackSamplesBuffer)
		}
		misc.BlockForSeconds(a.inputLoopTimeout)
	}
}

func (a *EchoAudio) outputCallbackLoop() {
	outputCallbackSampleBuffer := make([]Sample, a.outputBufferSize)
	for a.operate {
		a.outputCallback(outputCallbackSampleBuffer)

		a.sampleBufferLock.Lock()
		a.sampleBuffer = append(a.sampleBuffer, outputCallbackSampleBuffer...)
		a.sampleBufferLock.Unlock()

		misc.BlockForSeconds(a.outputLoopTimeout)
	}
}

func (a *EchoAudio) simulateChannel(samples []Sample) {
	maxBandwidth := a.outputSampleRate / 2
	simulatedBandWidth := a.highCut - a.lowCut
	equalizedNoiseLevel := a.noiseLevel * maxBandwidth / simulatedBandWidth
	for i, sample := range samples {
		noise := (rand.Float64()*2 - 1) * equalizedNoiseLevel
		newSample := (float64(sample) + noise) / (1 + equalizedNoiseLevel)
		newSample = a.bandPassFilter.Filter(newSample)
		samples[i] = Sample(newSample)
	}
}
