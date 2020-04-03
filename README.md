# transmission-jobs

[![Workflow](https://github.com/mark-ignacio/transmission-jobs/workflows/Go/badge.svg)](https://github.com/mark-ignacio/transmission-jobs/actions?query=workflow%3AGo)
[![GoDoc](https://godoc.org/github.com/mark-ignacio/transmission-jobs?status.svg)](https://godoc.org/github.com/mark-ignacio/transmission-jobs)

**transmission-jobs** is a job runner for Transmission RPC calls that supports torrent condition evaluation.

Runs every 5 minutes by default.

## Features

* [x] Condition evaluation
* [x] Tag system via jobs
* [x] RSS/Atom feed ingestion with filtering
* [x] Sonarr [History](https://github.com/Sonarr/Sonarr/wiki/History) integration
* [x] Actions
  * [x] Remove (and delete local data)

`transmission-jobs.default.yml` contains examples of feature usage.

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

### RSS and Atom feeds

RSS and Atom feeds are downloaded and processed each time transmission-jobs runs. If [stateful storage](#stateful-storage) is enabled, feed items are only created once.

You can optionally specify `feed.match`, which allows you to run an [RE2-compatible regular expression](https://github.com/google/re2/wiki/Syntax) against fields in a feed item. The full list of supported fields are `string` fields in [`gofeed.Item`](https://pkg.go.dev/github.com/mmcdole/gofeed?tab=doc#Item).

```yml
jobs:
  - name: feed pfSense amd64 ISOs
    feed:
      url: https://distrowatch.com/news/torrents.xml
      match:
        field: title
        regexp: pfSense\-.+?\-amd64
```

### Stateful storage

If `database` is configured, transmission-jobs changes its default stateless behavior to stateful. Other sections go into detail about what this means, but the affected job types are:

* `tag` - tags are stored after evaluated, which is generally useless
* `feed` - feed-added items are stored forever so that torrents are not added multiple times

Resetting storage is easy - just delete the file specified at `database` between transmission-jobs runs. 

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
