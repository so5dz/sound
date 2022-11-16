package config

type Config struct {
	Port    int        `json:"port"`
	Backend string     `json:"backend"`
	Input   Device     `json:"input"`
	Output  Device     `json:"output"`
	Echo    EchoConfig `json:"echo"`
	File    FileConfig `json:"file"`
}

type Device struct {
	Device     string `json:"device"`
	SampleRate int    `json:"sampleRate"`
	BufferSize int    `json:"bufferSize"`
}

type EchoConfig struct {
	NoiseLevel float64 `json:"noiseLevel"`
	LowCut     float64 `json:"lowCut"`
	HighCut    float64 `json:"highCut"`
	BPFOrder   int     `json:"bpfOrder"`
}

type FileConfig struct {
	Input string `json:"input"`
}
