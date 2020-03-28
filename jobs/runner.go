// Package jobs implements the transmissionrpc-powered job running system.
package jobs

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"log"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"
	"github.com/hekmon/transmissionrpc"
	"go.etcd.io/bbolt"
)

var (
	tagBucket = []byte("tags")
)

// Runner is what actually runs all of the jobs
type Runner struct {
	Config             Config
	DryRun             bool
	Verbose            bool
	client             *transmissionrpc.Client
	db                 *bbolt.DB
	sonarrDropPaths    map[string]bool
	allTorrents        []*TransmissionTorrent
	compiledConditions []*vm.Program
}

// Run runs the runner's configured jobs
func (r *Runner) Run(ctx context.Context) (err error) {
	// pop open the database
	if r.Config.DatabasePath != "" {
		r.db, err = bbolt.Open(r.Config.DatabasePath, 0600, nil)
		if err != nil {
			return
		}
		err = r.createBuckets()
		if err != nil {
			return
		}
		log.Printf("[*] Using database @ %s", r.Config.DatabasePath)
		defer r.db.Close()
	}
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
	for _, jobConfig := range r.Config.Jobs {
		log.Printf("[*] Running job: %s", jobConfig.Name)
		err = r.do(jobConfig)
		if err != nil {
			return fmt.Errorf("error running job '%s': %+v", jobConfig.Name, err)
		}
	}
	// store torrent data
	for _, torrent := range r.allTorrents {
		err = r.storeTorrent(torrent)
		if err != nil {
			return fmt.Errorf("error returning: %+v", err)
		}
	}
	return
}

func (r *Runner) do(job JobConfig) error {
	var err error
	if job.RemoveOptions != nil {
		if err = r.fetchAllTorrents(); err != nil {
			return err
		}
		err = r.remove(job)
	} else if job.TagOptions != nil {
		err = r.tag(job)
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
	r.allTorrents = make([]*TransmissionTorrent, len(allTorrents))
	for i := range allTorrents {
		torrent := ToTransmissionTorrent(*allTorrents[i], r.sonarrDropPaths)
		r.allTorrents[i] = &torrent
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
		return r.client.TorrentRemove(payload)
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
			r.allTorrents[i].Tags = append(r.allTorrents[i].Tags, tagName)
		}
	}
	return nil
}

func (r *Runner) createBuckets() error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		for _, name := range [][]byte{tagBucket} {
			_, err := tx.CreateBucketIfNotExists(name)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Runner) storeTorrent(torrent *TransmissionTorrent) error {
	key := make([]byte, 8)
	binary.PutVarint(key, torrent.ID)
	return r.db.Update(func(tx *bbolt.Tx) error {
		var (
			bucket = tx.Bucket(tagBucket)
			buf    = &bytes.Buffer{}
			enc    = gob.NewEncoder(buf)
		)
		enc.Encode(torrent.MarshalMap())
		return bucket.Put(key, buf.Bytes())
	})
}
