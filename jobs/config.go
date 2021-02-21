package jobs

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/mmcdole/gofeed"
)

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
	Location      string
	SeedRatio     float64        `mapstructure:"seed_ratio"`
	RemoveOptions *RemoveOptions `mapstructure:"remove"`
	TagOptions    *TagOptions    `mapstructure:"tag"`
	FeedOptions   *FeedOptions   `mapstructure:"feed"`
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
	Tag   string            // optional
	Match *FeedMatchOptions // optional
}

// FeedMatchOptions describes a regular expression to run on a particular feed field.
type FeedMatchOptions struct {
	Field  string
	RegExp string

	regexp *regexp.Regexp
}

var (
	validFeedFields = make(map[string]string)
)

// Validate returns whether this is a legit thing we can do or not (and caches some stuff)
func (f *FeedOptions) Validate() error {
	if f.Match != nil {
		_, valid := validFeedFields[strings.ToLower(f.Match.Field)]
		if !valid {
			return fmt.Errorf("invalid feed.match.field name: %s", f.Match.Field)
		}
		if f.Match.RegExp == "" {
			return fmt.Errorf("must specify feed.match.regexp")
		}
		compiled, err := regexp.Compile(f.Match.RegExp)
		if err != nil {
			return fmt.Errorf("invalid feed.match.regexp: %+v", err)
		}
		f.Match.regexp = compiled
	}
	return nil
}

func feedItemMatches(item gofeed.Item, options *FeedMatchOptions) bool {
	if options == nil {
		return true
	}
	field := reflect.ValueOf(item).FieldByName(validFeedFields[strings.ToLower(options.Field)])
	return options.regexp.MatchString(field.String())
}

func init() {
	itemType := reflect.TypeOf(gofeed.Item{})
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)
		if field.Type.Kind() == reflect.String {
			validFeedFields[strings.ToLower(field.Name)] = field.Name
		}
	}
}
