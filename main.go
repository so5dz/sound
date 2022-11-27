package main

import (
	"log"
	"strings"
	"unsafe"

	"github.com/so5dz/network/server"
	"github.com/so5dz/network/server/tcp"
	"github.com/so5dz/sound/audio"
	soundconfig "github.com/so5dz/sound/config"
	utils "github.com/so5dz/utils/config"
	"github.com/so5dz/utils/convert"
	"github.com/so5dz/utils/misc"
)

const _AudioSetupError = "unable to setup audio"

func main() {
	misc.WrapMain(mainWithError)()
}

func mainWithError() error {
	cfg, err := utils.LoadConfigFromArgs[soundconfig.Config]()
	if err != nil {
		return err
	}

	log.Println("Opening audio port")
	audioPort, err := getAudioPort(cfg)
	if err != nil {
		return misc.WrapError(_AudioSetupError, err)
	}

	log.Println("Opening TCP server on port", cfg.Port)
	tcpServer := tcp.StreamServer{}
	tcpServer.Initialize(cfg.Port)

	networkBuffer := convert.ByteFloatBuffer{}

	audioPort.OutputCallback(func(samples []audio.Sample) {
		networkSamples := networkBuffer.Get(cfg.Output.BufferSize)
		copy(asFloat64Array(samples), networkSamples)
		for i := len(networkSamples); i < len(samples); i++ {
			samples[i] = audio.Silence
		}
	})

	audioPort.InputCallback(func(samples []audio.Sample) {
		float64Samples := asFloat64Array(samples[0:cfg.Input.BufferSize])
		bytes := convert.FloatsToBytes(float64Samples)
		err := tcpServer.Broadcast(bytes)
		if err != nil {
			log.Println(err)
		}
	})

	tcpServer.OnReceive(func(r server.Remote, b []byte) {
		networkBuffer.Put(b)
	})

	log.Println("Initializing audio port")
	audioPort.Initialize(cfg)

	log.Println("Opening audio port")
	audioPort.Open()

	log.Println("Starting TCP server")
	tcpServer.Start()

	log.Println("MARDES-sound started, interrupt to close")
	misc.BlockUntilInterrupted()

	tcpServer.Stop()
	audioPort.Close()
	audioPort.Dispose()

	log.Println("Closing threads")
	misc.BlockForSeconds(1)

	log.Println("Closing")
	return nil
}

func getAudioPort(cfg soundconfig.Config) (audio.Audio, error) {
	backend := strings.ToUpper(cfg.Backend)

	switch backend {
	case "PORTAUDIO":
		ap := &audio.PortAudio{}
		return ap, nil

	case "FILE":
		ap := &audio.FileAudio{}
		return ap, nil

	case "ECHO":
		break
	}

	ap := &audio.EchoAudio{}
	return ap, nil
}

func asFloat64Array(a []audio.Sample) []float64 {
	return (*(*[]float64)(unsafe.Pointer(&a)))[:]
}
