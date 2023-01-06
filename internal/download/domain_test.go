package download_test

import (
	"github.com/MarcoTomasRodriguez/hget/internal/download"
	"github.com/stretchr/testify/assert"
	"testing"
)

var golangSample = download.Download{
	Id:       "v5pra7bt",
	Name:     "go1.19.1.src.tar.gz",
	URL:      "https://go.dev/dl/go1.19.1.src.tar.gz",
	Size:     1300,
	Segments: []download.Segment{{"v5pra7bt/segment.00", 0, 1300}},
}

var javaSample = download.Download{
	Id:   "ita2qybt",
	Name: "jre-8u351-macosx-x64.dmg",
	URL:  "https://java.com/download/jre/jre-8u351-macosx-x64.dmg",
	Size: 2583,
	Segments: []download.Segment{
		{"ita2qybt/segment.00", 0, 644},
		{"ita2qybt/segment.01", 645, 1289},
		{"ita2qybt/segment.02", 1290, 1934},
		{"ita2qybt/segment.03", 1935, 2583},
	},
}

func TestDownload_String(t *testing.T) {
	assert.Equal(t, " ⁕ v5pra7bt ⇒ URL: https://go.dev/dl/go1.19.1.src.tar.gz Size: 1.3 kB\n", golangSample.String())
}
