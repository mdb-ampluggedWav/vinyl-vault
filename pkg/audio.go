package pkg

type Bitrate int

const (
	BitrateUnknown Bitrate = iota
	BitrateLow
	BitrateMedium
	BitrateHigh
	BitrateCDQuality
	BitrateHi_Res
)

func (b Bitrate) Kbps() int {

	switch b {
	case BitrateLow:
		return 128
	case BitrateMedium:
		return 192
	case BitrateHigh:
		return 256
	case BitrateCDQuality:
		return 1411
	case BitrateHi_Res:
		return 2034
	default:
		return 0
	}
}

type PlayMode int

const (
	PlayModeStopped PlayMode = iota
	PlayModePlaying
	PlayModePaused
)

func (pm PlayMode) String() string {
	switch pm {
	case PlayModePlaying:
		return "playing"
	case PlayModePaused:
		return "paused"
	case PlayModeStopped:
		return "stopped"
	default:
		return "unknown"
	}
}

type Compression int

const (
	CompressionUnknown Compression = iota
	CompressionNone
	CompressionLossy
	CompressionLossless
)

type Format struct {
	Name        string
	Compression Compression
}

var audioFormats = map[string]Format{
	"WAV":  {Name: "WAV", Compression: CompressionNone},
	"AIFF": {Name: "AIFF", Compression: CompressionNone},
	"FLAC": {Name: "FLAC", Compression: CompressionLossless},
	"ALAC": {Name: "Apple Lossless", Compression: CompressionLossless},
	"MP3":  {Name: "MP3", Compression: CompressionLossy},
}

func GetAudioFormats() map[string]Format {
	return audioFormats
}

type AudioQuality struct {
	Bitrate Bitrate `json:"bitrate"`
	Format  string  `json:"format"`
}
