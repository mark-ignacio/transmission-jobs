package jobs

import (
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
	if t.sonarrImportPaths == nil {
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
		_, exists := t.sonarrImportPaths[filePath]
		if !exists {
			return false
		}
	}
	return true
}

func init() {
	torrentExprEnv = expr.Env(&torrentConditionInput{})
}
