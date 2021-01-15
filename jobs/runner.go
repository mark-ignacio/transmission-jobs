// Package jobs implements the transmissionrpc-powered job running system.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/timshannon/bolthold"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/hekmon/transmissionrpc"
	"github.com/mmcdole/gofeed"
)

var (
	feedGUIDBucket = []byte("feedGUIDs")
	torrentsBucket = []byte("torrents")
)

// Runner is what actually runs all of the jobs
type Runner struct {
	Config             Config
	DryRun             bool
	Verbose            bool
	client             *transmissionrpc.Client
	db                 *bolthold.Store
	sonarrDropPaths    map[string]bool
	allTorrents        map[int64]*TransmissionTorrent
	compiledConditions []*vm.Program
	feedCache          map[string]*gofeed.Feed
}

// Run runs the runner's configured jobs
func (r *Runner) Run(ctx context.Context) (err error) {
	// pop open the database
	if r.Config.DatabasePath != "" {
		r.db, err = bolthold.Open(r.Config.DatabasePath, 0600, nil)
		if err != nil {
			return
		}
		log.Printf("[*] Using database @ %s", r.Config.DatabasePath)
		defer r.db.Close()
	}
	r.allTorrents = make(map[int64]*TransmissionTorrent)
	r.feedCache = make(map[string]*gofeed.Feed)
	// validate jobs before we do any network stuff
	if err = r.validateJobs(); err != nil {
		return
	}
	r.client, err = ConnectToRemote(r.Config.Transmission)
	if err != nil {
		return
	}
	if r.Config.Sonarr != nil {
		r.sonarrDropPaths, err = FetchSonarrDrops(*r.Config.Sonarr, 1000)
		if err != nil {
			return
		}
	}
	if r.DryRun {
		log.Println("[*] Dry run mode - no changes will be made")
	}
	// load it up!
	err = r.fetchAllTorrents()
	if err != nil {
		return fmt.Errorf("could not perform initial fetch of all torrents: %+v", err)
	}
	if r.db != nil {
		err = r.loadTorrentStates()
		if err != nil {
			return fmt.Errorf("error loading saved torrent states: %+v", err)
		}
	}
	for _, jobConfig := range r.Config.Jobs {
		log.Printf("[*] Running job: %s", jobConfig.Name)
		err = r.do(jobConfig)
		if err != nil {
			return fmt.Errorf("error running job '%s': %+v", jobConfig.Name, err)
		}
	}
	if r.db != nil {
		// store torrent data
		for _, torrent := range r.allTorrents {
			err = r.storeTorrent(torrent)
			if err != nil {
				return fmt.Errorf("error storing torrent: %+v", err)
			}
		}
	}
	return
}

func (r *Runner) do(job JobConfig) error {
	var err error
	if job.RemoveOptions != nil {
		err = r.remove(job)
	} else if job.TagOptions != nil {
		err = r.tag(job)
	} else if job.FeedOptions != nil {
		err = r.feed(job)
	} else {
		err = fmt.Errorf("invalid job spec for %s", job.Name)
	}
	return err
}

func (r *Runner) fetchAllTorrents() error {
	if r.Verbose {
		log.Println("[*] Getting all torrents...")
	}
	allTorrents, err := r.client.TorrentGetAll()
	if err != nil {
		return fmt.Errorf("error getting all torrents: %+v", err)
	}
	for i := range allTorrents {
		torrent := ToTransmissionTorrent(*allTorrents[i], r.sonarrDropPaths)
		r.allTorrents[torrent.ID] = &torrent
	}
	return nil
}

func (r *Runner) validateJobs() error {
	r.compiledConditions = make([]*vm.Program, len(r.Config.Jobs))
	for i, jobConfig := range r.Config.Jobs {
		err := r.validateJob(i, jobConfig)
		if err != nil {
			return fmt.Errorf("invalid job '%s': %+v", jobConfig.Name, err)
		}
	}
	return nil
}

func (r *Runner) validateJob(index int, job JobConfig) error {
	// at the moment, validation is just compiling conditions
	program := r.compiledConditions[index]
	if program != nil {
		return nil
	}
	var conditionStr string
	if job.RemoveOptions != nil {
		conditionStr = job.RemoveOptions.Condition
	} else if job.TagOptions != nil {
		conditionStr = job.TagOptions.Condition
	} else if job.FeedOptions != nil {
		return job.FeedOptions.Validate()
	}
	program, err := expr.Compile(conditionStr, torrentExprEnv)
	if err != nil {
		return fmt.Errorf("error compiling condition '%s':\n%+v", conditionStr, err)
	}
	r.compiledConditions[index] = program
	return nil
}

// TODO: refactor and move all of these out of the struct?
func (r *Runner) remove(job JobConfig) error {
	// validate condition
	if job.RemoveOptions == nil || job.RemoveOptions.Condition == "" {
		return errors.New("job has invalid RemoveOptions")
	}
	conditionStr := job.RemoveOptions.Condition
	conditionProgram, err := expr.Compile(conditionStr, torrentExprEnv)
	if err != nil {
		return fmt.Errorf("error compiling condition '%s':\n%+v", conditionStr, err)
	}
	removeIDs := []int64{}
	for _, torrent := range r.allTorrents {
		output, err := expr.Run(conditionProgram, &torrentConditionInput{*torrent})
		if err != nil {
			return fmt.Errorf("error evaluting condition '%s':\n:%+v", conditionStr, err)
		}
		if output.(bool) {
			if r.DryRun {
				log.Printf("DRY RUN: remove %s", torrent.Name)
			} else {
				if r.Verbose {
					log.Printf("queueing %s for removal", torrent.Name)
				}
				removeIDs = append(removeIDs, torrent.ID)
			}
		}
	}
	if len(removeIDs) > 0 {
		payload := &transmissionrpc.TorrentRemovePayload{
			IDs:             removeIDs,
			DeleteLocalData: job.RemoveOptions.DeleteLocal,
		}
		log.Printf("[+] Removing %d torrents", len(removeIDs))
		if r.Verbose {
			log.Printf("[*] removing IDs: %v", removeIDs)
		}
		err := r.client.TorrentRemove(payload)
		if err != nil {
			return err
		}
		for _, id := range removeIDs {
			storedInfo := r.allTorrents[id].StoredTorrentInfo
			storedInfo.Removed = true
			if r.db != nil && storedInfo.SafeToPrune() {
				if r.DryRun {
					log.Printf("DRY RUN: would prune info")
				} else if err := r.db.Delete(id, storedInfo); err != nil {
					return fmt.Errorf("error deleting stored torrent info for ID %d: %+v", id, err)
				}
			} else {
				if r.DryRun {
					log.Printf("DRY RUN: would not prune info")
				} else if err = r.db.Update(storedInfo.ID, storedInfo); err != nil {
					return fmt.Errorf("error saving storted torrent info for ID %d: %+v", id, err)
				}
			}
			delete(r.allTorrents, id)
		}
	}
	return nil
}

func (r *Runner) tag(job JobConfig) error {
	// validate condition
	if job.TagOptions == nil || job.TagOptions.Condition == "" {
		return errors.New("job has invalid RemoveOptions")
	}
	conditionStr := job.TagOptions.Condition
	conditionProgram, err := expr.Compile(conditionStr, torrentExprEnv)
	if err != nil {
		return fmt.Errorf("error compiling condition '%s':\n%+v", conditionStr, err)
	}
	for i, torrent := range r.allTorrents {
		output, err := expr.Run(conditionProgram, &torrentConditionInput{*torrent})
		if err != nil {
			return fmt.Errorf("error evaluting condition '%s':\n:%+v", conditionStr, err)
		}
		if output.(bool) {
			// tags don't mutate state, so no dry run necessary
			tagName := job.TagOptions.Name
			if r.Verbose {
				log.Printf("[*] Tagging %s with '%s'", torrent.Name, tagName)
			}
			stored := r.allTorrents[i].GetOrCreateStored()
			r.allTorrents[i].Tags = append(stored.Tags, tagName)
		}
	}
	return nil
}

func (r *Runner) feed(job JobConfig) error {
	if job.FeedOptions.URL == "" {
		return fmt.Errorf("feed job does not have a URL")
	}
	var (
		err          error
		parser       = gofeed.NewParser()
		feed, exists = r.feedCache[job.FeedOptions.URL]
	)
	if !exists {
		feed, err = parser.ParseURL(job.FeedOptions.URL)
		if err != nil {
			return err
		}
		r.feedCache[job.FeedOptions.URL] = feed
	}
	for _, item := range feed.Items {
		if item.Link == "" {
			log.Printf("[*] %s item does not have a Link, skipping", job.FeedOptions.URL)
		}
		// if we enabled storage, check if we already downloaded it
		if r.db != nil {
			// item.GUID
			var stored StoredTorrentInfo
			err = r.db.FindOne(&stored, bolthold.Where("FeedGUID").Eq(item.GUID).Index("FeedGUID"))
			if err == bolthold.ErrNotFound {
				if !feedItemMatches(*item, job.FeedOptions.Match) {
					continue
				}
			} else if err != nil {
				return fmt.Errorf("error checking if %s is downloaded: %+v", item.GUID, err)
			} else {
				// already exists or item does not match
				continue
			}
		}
		if r.DryRun {
			log.Printf("DRY RUN: would add feed item: %s (%s)", item.Title, item.Link)
			continue
		}
		var downloadDir *string
		if job.Location != "" {
			downloadDir = &job.Location
		}
		log.Printf("[*] Adding %s", item.Title)
		torrent, err := r.client.TorrentAdd(
			&transmissionrpc.TorrentAddPayload{
				Filename:    &item.Link,
				DownloadDir: downloadDir,
			},
		)
		if err != nil {
			return fmt.Errorf("unable to add torrent from feed item %s: %+v", item.GUID, err)
		}
		var (
			transTorrent = TransmissionTorrent{
				ID:         *torrent.ID,
				Name:       *torrent.Name,
				HashString: *torrent.HashString,
			}
			stored = transTorrent.GetOrCreateStored()
		)
		stored.FeedGUID = item.GUID
		if job.FeedOptions.Tag != "" {
			stored.Tags = append(stored.Tags, job.FeedOptions.Tag)
		}
		if r.db != nil {
			r.db.Upsert(stored.ID, stored)
		}
	}
	return nil
}

func (r *Runner) storeTorrent(torrent *TransmissionTorrent) error {
	if torrent.StoredTorrentInfo == nil {
		return nil
	}
	return r.db.Upsert(torrent.ID, torrent.StoredTorrentInfo)
}

func (r *Runner) loadTorrentStates() error {
	var toRemove []int64
	err := r.db.ForEach(nil, func(info *StoredTorrentInfo) error {
		_, exists := r.allTorrents[info.ID]
		if exists {
			r.allTorrents[info.ID].StoredTorrentInfo = info
		} else if info.SafeToPrune() {
			toRemove = append(toRemove, info.ID)
		}
		return nil
	})
	for id, torrent := range r.allTorrents {
		if torrent.StoredTorrentInfo == nil {
			r.allTorrents[id].GetOrCreateStored()
		}
	}
	return err
}
