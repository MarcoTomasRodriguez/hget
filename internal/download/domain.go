package download

import (
	"fmt"
	"github.com/MarcoTomasRodriguez/hget/pkg/fsutil"
	"github.com/fatih/color"
)

// Download stores the information of a resource that can be downloaded.
type Download struct {
	Id       string    `yaml:"id"`
	Name     string    `yaml:"name"`
	URL      string    `yaml:"url"`
	Size     int64     `yaml:"size"`
	Segments []Segment `yaml:"segments"`
}

// Segment stores the start and end points of a download's segment.
type Segment struct {
	// Format: {downloadId}/{segmentId}
	// Example: a1b2c3d4/segment.01
	Id    string `yaml:"id"`
	Start int64  `yaml:"start"`
	End   int64  `yaml:"end"`
}

// String returns a colored formatted string with the _download's Id, URL and Size.
func (d Download) String() string {
	return fmt.Sprintln(
		" ⁕", color.HiCyanString(d.Id), "⇒",
		color.HiCyanString("URL:"), d.URL,
		color.HiCyanString("Size:"), fsutil.ReadableMemorySize(d.Size),
	)
}
