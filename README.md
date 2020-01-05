# transmission-jobs

[![Workflow](https://github.com/mark-ignacio/transmission-jobs/workflows/Go/badge.svg)](https://github.com/mark-ignacio/transmission-jobs/actions?query=workflow%3AGo)
[![GoDoc](https://godoc.org/github.com/mark-ignacio/transmission-jobs?status.svg)](https://godoc.org/github.com/mark-ignacio/transmission-jobs)

**transmission-jobs** is a job runner for Transmission RPC calls that supports torrent condition evaluation.

Runs every 5 minutes by default.

## Features

* [x] Condition evaluation
* [x] Ephemeral tag system via jobs
* [x] Sonarr [History](https://github.com/Sonarr/Sonarr/wiki/History) integration
* [x] Actions
  * [x] Remove (and delete local data)

### Conditions

Conditions use <https://github.com/antonmedv/expr/> as the boolean expression engine. All conditions are validated before jobs are run, so you should get informative error messages before bad things happen on runtime.

Conditions have a `Torrent` variable, which is the [TransmissionTorrent](https://godoc.org/github.com/mark-ignacio/transmission-jobs/jobs#TransmissionTorrent) object in question:

```yaml
jobs:
- name: tag Fedora, Debian trackers as linux
  tag:
    name: linux
    condition: |-
      any(Torrent.AnnounceHostnames(), {# in ["torrent.fedoraproject.org", "bttracker.debian.org"]})
  - name: delete non-linux + seeding where ratio > 10
    remove:
      condition: "linux" not in Torrent.Tags && Torrent.Status.String() == "seeding" && Torrent.UploadRatio >= 10.0
      delete_local: true
```

Errors are straightforward:

```sh
# for `condition: Torrent.Staaatus == "seeding"`
$ go run main.go -n
2020/01/01 10:15:33 error running jobs: invalid job 'garbage': error compiling condition 'Torrent.Staaatus == "seeding"':
type jobs.TransmissionTorrent has no field Staaatus (1:1)
 | Torrent.Staaatus == "seeding"
 | ^
exit status 1

# for `condition: Torrent.Status == "seeding"`
2020/01/01 10:17:07 error running jobs: invalid job 'garbage': error compiling condition 'Torrent.Status == "seeding"':
invalid operation: == (mismatched types transmissionrpc.TorrentStatus and string) (1:16)
 | Torrent.Status == "seeding"
 | ...............^
exit status 1
```

See the expr [Language Definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md) for details.

### Sonarr import status

Optionally specifying Sonarr connection information allows calling [`Torrent.Imported()`](https://godoc.org/github.com/mark-ignacio/transmission-jobs/jobs#TransmissionTorrent.Imported) inside of conditions:

```yaml
sonarr:
  host: https://localhost:8989
  api_key: deadbeef
jobs:
  - name: delete imported + seeding + ratio > 10
    remove:
      condition: Torrent.Imported() && Torrent.Status.String() == "seeding" && Torrent.UploadRatio >= 10.0
      delete_local: true
```
