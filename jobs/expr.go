package jobs

import (
	"log"
	"net/url"
	"path"

	"github.com/antonmedv/expr"
)

var (
	torrentExprEnv expr.Option
)

type torrentConditionInput struct {
	Torrent TransmissionTorrent
}

// Imported returns whether all downloaded files were imported
func (t TransmissionTorrent) Imported() bool {
	if t.sonarrDropPaths == nil {
		panic("runtime error: unable to use Imported() without loading imported file paths")
	}
	if len(t.Files) == 0 {
		return false
	}
	for _, fileData := range t.Files {
		filePath := path.Join(
			t.DownloadDir,
			fileData.Name,
		)
		_, exists := t.sonarrDropPaths[filePath]
		if !exists {
			return false
		}
	}
	return true
}

// AnnounceHostnames returns a list of tracker announce URL hostnames.
func (t TransmissionTorrent) AnnounceHostnames() (hostnames []string) {
	if len(t.Trackers) == 0 {
		return
	}
	for _, tracker := range t.Trackers {
		trackerURL, err := url.Parse(tracker.Announce)
		if err != nil {
			log.Printf("could not parse announce URL '%s' as a URL: %+v", tracker.Announce, err)
			continue
		}
		hostnames = append(hostnames, trackerURL.Hostname())
	}
	return
}

func init() {
	torrentExprEnv = expr.Env(&torrentConditionInput{})
}
