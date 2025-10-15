package pkg

import (
	"fmt"
	"time"
)

type Metadata struct {
	Artist       string   `json:"artist" gorm:"column:artist"`
	Album        string   `json:"album" gorm:"column:album"`
	Format       string   `json:"format" gorm:"column:format"`
	ReleaseDate  string   `json:"release_date" gorm:"column:release_date"`
	Label        *string  `json:"label,omitempty" gorm:"column:label"`
	Country      *string  `json:"country,omitempty" gorm:"column:country"`
	Length       Duration `json:"length" gorm:"column:length"`
	CoverArtPath string   `json:"cover_art_path" gorm:"column:cover_art_path"`
}

type Duration uint

func (d Duration) ToTime() time.Duration {
	return time.Duration(d) * time.Second
}

func (d Duration) String() string {
	dur := d.ToTime()
	minutes := int(dur.Minutes())
	seconds := int(dur.Seconds()) % 60
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

func (d Duration) StringLong() string {
	dur := d.ToTime()
	hours := int(dur.Hours())
	minutes := int(dur.Minutes()) % 60
	seconds := int(dur.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}
