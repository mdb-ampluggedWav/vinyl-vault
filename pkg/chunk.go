package pkg

type ChunkManifest struct {
	TrackID     uint64      `json:"track_id"`
	Filename    string      `json:"filename"`
	TotalSize   int64       `json:"total_size"`
	ChunkSize   int64       `json:"chunk_size"`
	TotalChunks int         `json:"total_chunks"`
	Checksum    string      `json:"checksum"` // SHA-256 of full file
	Chunks      []ChunkInfo `json:"chunks"`
}

type ChunkInfo struct {
	Index    int    `json:"index"`
	Offset   int64  `json:"offset"`
	Size     int64  `json:"size"`
	Checksum string `json:"checksum"` // SHA-256 of a specific chunk
}

type DownloadProgress struct {
	TrackID         uint64  `json:"track_id"`
	CompletedChunks []int   `json:"completed_chunks"`
	TotalChunks     int     `json:"total_chunks"`
	Percentage      float64 `json:"percentage"`
}
