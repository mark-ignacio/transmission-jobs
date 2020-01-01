# transmission-jobs

**transmission-jobs** is a job runner for Transmission RPC calls that supports torrent condition evaluation.

Runs every 5 minutes by default.

## Features

* [x] Condition evaluation
* [ ] Ephemeral tag system (via a job?)
* [ ] Sonarr [EpisodeFile](https://github.com/Sonarr/Sonarr/wiki/EpisodeFile) integration
* Actions
  * [x] Remove (and delete local data)

## Writing conditions

Conditions use <https://github.com/antonmedv/expr/> as the expression engine. The `Torrent` struct is a `TransmissionTorrent

See the [Language Definition](https://github.com/antonmedv/expr/blob/master/docs/Language-Definition.md) for details.
