package jobs

// Config describes the schema of the .yml config file
type Config struct {
	DatabasePath string `mapstructure:"database"`
	Transmission TransmissionSettings
	Sonarr       *SonarrSettings
	Jobs         []JobConfig
}

// SonarrSettings describes how to connect to a Sonarr server.
type SonarrSettings struct {
	Host   string
	APIKey string `mapstructure:"api_key"`
}

// TransmissionSettings describes how to connect to a Transmission RPC server.
type TransmissionSettings struct {
	Host     string
	Username string
	Password string
}

// JobConfig describes jobs to run. The presence of each 'SomethingOptions' field denotes the action.
type JobConfig struct {
	Name          string
	RemoveOptions *RemoveOptions `mapstructure:"remove"`
	TagOptions    *TagOptions    `mapstructure:"tag"`
	FeedOptions   *FeedOptions   `mapstructure:"rss"`
}

// RemoveOptions describes when and how to remove a torrent.
type RemoveOptions struct {
	DeleteLocal bool `mapstructure:"delete_local"`
	Condition   string
}

// TagOptions describes when and how to tag a torrent.
type TagOptions struct {
	Name      string
	Condition string
	Ephemeral bool
}

// FeedOptions describes how to add a torrent from an Atom/RSS feed.
type FeedOptions struct {
	URL   string
	Match *FeedMatchOptions //optional
}

// FeedMatchOptions describes a regular expression to run on a particular feed field.
type FeedMatchOptions struct {
	Field  string
	RegExp string
}
