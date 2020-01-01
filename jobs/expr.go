package jobs

import (
	"github.com/antonmedv/expr"
)

var (
	torrentExprEnv expr.Option
)

type torrentConditionInput struct {
	Torrent TransmissionTorrent
}

func init() {
	torrentExprEnv = expr.Env(&torrentConditionInput{})
}
