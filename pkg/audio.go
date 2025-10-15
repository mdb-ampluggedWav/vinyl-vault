package pkg

type AudioQuality struct {
	Format     string `json:"format"`
	Bitrate    int    `json:"bitrate"`
	SampleRate int    `json:"sample_rate"`
	BitDepth   int    `json:"bitDepth"`
	Channels   int    `json:"channels"`
}
