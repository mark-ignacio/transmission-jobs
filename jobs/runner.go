// Package jobs implements the transmissionrpc-powered job running system.
package jobs

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/antonmedv/expr"
	"github.com/antonmedv/expr/vm"

	"github.com/hekmon/transmissionrpc"
)

// Runner is what actually runs all of the jobs
type Runner struct {
	Config             Config
	DryRun             bool
	Verbose            bool
	client             *transmissionrpc.Client
	allTorrents        []TransmissionTorrent
	compiledConditions []*vm.Program
}

// Run runs the runner's configured jobs
func (r *Runner) Run(ctx context.Context) (err error) {
	// validate jobs before we do any network stuff
	if err = r.validateJobs(); err != nil {
		return
	}
	r.client, err = ConnectToRemote(r.Config.Transmission)
	if err != nil {
		return
	}
	if r.DryRun {
		log.Println("[*] Dry run mode - no changes will be made")
	}
	for _, jobConfig := range r.Config.Jobs {
		log.Printf("[*] Running job: %s", jobConfig.Name)
		err = r.do(jobConfig)
		if err != nil {
			return fmt.Errorf("error running job '%s': %+v", jobConfig.Name, err)
		}
	}
	return
}

func (r *Runner) do(job JobConfig) error {
	if r.Verbose {
		log.Println("[*] Getting all torrents...")
	}
	allTorrents, err := r.client.TorrentGetAll()
	if err != nil {
		return fmt.Errorf("error getting all torrents: %+v", err)
	}
	r.allTorrents = make([]TransmissionTorrent, len(allTorrents))
	for i := range allTorrents {
		r.allTorrents[i] = ToTransmissionTorrent(*allTorrents[i])
	}
	if job.RemoveOptions != nil {
		err = r.remove(job)
	} else {

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
	conditionStr := job.RemoveOptions.Condition
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
		output, err := expr.Run(conditionProgram, &torrentConditionInput{torrent})
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
