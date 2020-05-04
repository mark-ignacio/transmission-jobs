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
		panic("runtime error: unable to use Imported() without loading imported Sonarr file paths")
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

// GetOrCreateStored gets or creates StoredTorrent info.
func (t *TransmissionTorrent) GetOrCreateStored() *StoredTorrentInfo {
	if t.StoredTorrentInfo == nil {
		t.StoredTorrentInfo = &StoredTorrentInfo{ID: t.ID}
	}
	return t.StoredTorrentInfo
}

// StoredTorrentInfo contains TransmissionTorrent info saved with bolthold. Usually embedded inside of TransmissionTorrent.
type StoredTorrentInfo struct {
	ID       int64  `boltholdKey:"ID"`
	FeedGUID string `boltholdIndex:"FeedGUID"`
	Removed  bool
	Tags     []string
}

// SafeToPrune returns whether a piece of stored torrent state is worth keeping around.
func (s StoredTorrentInfo) SafeToPrune() bool {
	if s.Removed && s.FeedGUID != "" {
		return false
	}
	return true
}

func init() {
	torrentExprEnv = expr.Env(&torrentConditionInput{})
}
