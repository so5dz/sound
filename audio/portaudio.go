package audio

import (
	soundconfig "github.com/iskrapw/sound/config"
	"github.com/iskrapw/utils/misc"

	"github.com/gordonklaus/portaudio"
)

const _MonoChannels = 1

const _DeviceListError = "unable to get device list"
const _InputStreamParametersBuildError = "unable to build input stream parameters"
const _OutputStreamParametersBuildError = "unable to build output stream parameters"
const _InputStreamOpenError = "unable to open input stream"
const _OutputStreamOpenError = "unable to open output stream"

type _RawSample float32

const _RawSilence _RawSample = 0

type PortAudio struct {
	inputStreamParameters  portaudio.StreamParameters
	outputStreamParameters portaudio.StreamParameters
	inputStreamCallback    func([]Sample)
	outputStreamCallback   func([]Sample)
	inputStream            *portaudio.Stream
	outputStream           *portaudio.Stream
}

func (a *PortAudio) Initialize(config soundconfig.Config) error {
	var err error
	portaudio.Initialize()

	a.inputStreamParameters, err = getStreamParameters(config.Input, false)
	if err != nil {
		return misc.WrapError(_InputStreamParametersBuildError, err)
	}

	a.outputStreamParameters, err = getStreamParameters(config.Output, true)
	if err != nil {
		return misc.WrapError(_OutputStreamParametersBuildError, err)
	}

	return nil
}

func (a *PortAudio) Dispose() {
	portaudio.Terminate()
}

func (a *PortAudio) Open() error {
	var err error

	a.inputStream, err = portaudio.OpenStream(a.inputStreamParameters, func(rawSamples []_RawSample) {
		if a.inputStreamCallback != nil {
			samples := asSamples(rawSamples)
			a.inputStreamCallback(samples)
		}
	})
	if err != nil {
		return misc.WrapError(_InputStreamOpenError, err)
	}

	a.outputStream, err = portaudio.OpenStream(a.outputStreamParameters, func(rawSamples []_RawSample) {
		if a.outputStreamCallback != nil {
			samples := asSamples(rawSamples)
			a.outputStreamCallback(samples)
			toRawSamples(rawSamples, samples)
		} else {
			for i := 0; i < a.outputStreamParameters.FramesPerBuffer; i++ {
				rawSamples[i] = _RawSilence
			}
		}
	})
	if err != nil {
		return misc.WrapError(_OutputStreamOpenError, err)
	}

	a.inputStream.Start()
	a.outputStream.Start()
	return nil
}

func (a *PortAudio) Close() {
	a.inputStream.Close()
	a.outputStream.Close()
}

func (a *PortAudio) InputCallback(c func([]Sample)) {
	a.inputStreamCallback = c
}

func (a *PortAudio) OutputCallback(c func([]Sample)) {
	a.outputStreamCallback = c
}

func getStreamParameters(deviceConfig soundconfig.Device, isOutput bool) (portaudio.StreamParameters, error) {
	var streamParameters portaudio.StreamParameters

	devices, err := portaudio.Devices()
	if err != nil {
		return streamParameters, misc.WrapError(_DeviceListError, err)
	}

	var selectedDevice *portaudio.DeviceInfo
	for _, device := range devices {
		if device.Name == deviceConfig.Device {
			selectedDevice = device
			break
		}
	}
	if selectedDevice == nil {
		return streamParameters, misc.NewError("device", deviceConfig.Device, "not found")
	}

	if isOutput {
		streamParameters = portaudio.LowLatencyParameters(nil, selectedDevice)
		streamParameters.Output.Channels = _MonoChannels
	} else {
		streamParameters = portaudio.LowLatencyParameters(selectedDevice, nil)
		streamParameters.Input.Channels = _MonoChannels
	}
	streamParameters.SampleRate = float64(deviceConfig.SampleRate)
	streamParameters.FramesPerBuffer = deviceConfig.BufferSize

	return streamParameters, nil
}

func toRawSamples(destination []_RawSample, samples []Sample) {
	for i, f := range samples {
		destination[i] = _RawSample(f)
	}
}

func asSamples(rawSamples []_RawSample) []Sample {
	samples := make([]Sample, len(rawSamples))
	for i, f := range rawSamples {
		samples[i] = Sample(f)
	}
	return samples
}
