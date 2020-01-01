package jobs

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/hekmon/transmissionrpc"
)

// Config describes the schema of the .yml config file
type Config struct {
	Transmission TransmissionSettings
	Jobs         []JobConfig
}

// TransmissionSettings describes how to connect to a Transmission RPC endpoint.
type TransmissionSettings struct {
	Host     string
	Username string
	Password string
}

// JobConfig describes jobs to run. The presence of each 'SomethingOptions' field denotes the action.
type JobConfig struct {
	Name          string
	RemoveOptions *RemoveOptions `mapstructure:"remove"`
}

// RemoveOptions describes when and how to remove a torrent.
type RemoveOptions struct {
	DeleteLocal bool `mapstructure:"delete_local"`
	Condition   string
}

// ConnectToRemote creates an *transmissionrpc.Client.
func ConnectToRemote(settings TransmissionSettings) (*transmissionrpc.Client, error) {
	uri, err := url.Parse(settings.Host)
	if err != nil {
		return nil, fmt.Errorf("error parsing Host: %+v", err)
	}
	port, err := strconv.ParseUint(uri.Port(), 10, 16)
	if err != nil {
		return nil, fmt.Errorf("error parsing port as uint16: %+v", err)
	}
	advancedConfig := &transmissionrpc.AdvancedConfig{
		Port:  uint16(port),
		HTTPS: (strings.ToLower(uri.Scheme) == "https"),
	}
	return transmissionrpc.New(
		uri.Hostname(),
		settings.Username,
		settings.Password,
		advancedConfig,
	)
}
