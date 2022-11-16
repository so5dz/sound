package audio

import (
	"encoding/binary"
	"io/ioutil"
	"math"

	soundconfig "github.com/iskrapw/sound/config"
	"github.com/iskrapw/utils/misc"
)

type FileAudio struct {
	operate              bool
	sampleBuffer         []Sample
	sampleBufferPosition int

	inputCallback    func([]Sample)
	inputBufferSize  int
	inputLoopTimeout float64

	sampleRate float64
}

func (a *FileAudio) Initialize(config soundconfig.Config) error {
	a.operate = false
	a.sampleRate = float64(config.Input.SampleRate)

	err := a.loadFileIntoBuffer(config.File.Input)
	if err != nil {
		return misc.WrapError("unable to open specified file", err)
	}

	a.inputBufferSize = config.Input.BufferSize
	a.inputLoopTimeout = float64(a.inputBufferSize) / float64(config.Input.SampleRate)
	return nil
}

func (a *FileAudio) Dispose() {

}

func (a *FileAudio) Open() error {
	a.operate = true
	go a.inputCallbackLoop()
	return nil
}

func (a *FileAudio) Close() {
	a.operate = false
}

func (a *FileAudio) InputCallback(c func([]Sample)) {
	a.inputCallback = func(samples []Sample) {
		c(samples)
	}
}

func (a *FileAudio) OutputCallback(c func([]Sample)) {

}

func (a *FileAudio) inputCallbackLoop() {
	callbackSlice := make([]Sample, a.inputBufferSize)
	for a.operate {
		fill := 0
		for fill < a.inputBufferSize {
			toEnd := len(a.sampleBuffer) - a.sampleBufferPosition
			if toEnd >= a.inputBufferSize {
				copy(callbackSlice[fill:], a.sampleBuffer[a.sampleBufferPosition:])
				fill += a.inputBufferSize
				a.sampleBufferPosition += a.inputBufferSize
			} else {
				copy(callbackSlice[fill:], a.sampleBuffer[a.sampleBufferPosition:])
				fill += toEnd
				a.sampleBufferPosition = toEnd
			}
		}

		a.inputCallback(callbackSlice)
		misc.BlockForSeconds(a.inputLoopTimeout)
	}
}

func (a *FileAudio) loadFileIntoBuffer(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	const sampleSize = 4
	a.sampleBuffer = make([]Sample, len(content)/sampleSize)
	for i := 0; i < len(a.sampleBuffer); i++ {
		a.sampleBuffer[i] = Sample(math.Float32frombits(binary.LittleEndian.Uint32(content[sampleSize*i : sampleSize*(i+1)])))
	}

	a.sampleBufferPosition = 0
	return nil
}
