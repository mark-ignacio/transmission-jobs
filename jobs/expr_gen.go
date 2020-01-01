
package jobs

// *********************************************
// Code generated by gen/expr.go -- DO NOT EDIT.
// *********************************************

import (
	"time"

	"github.com/hekmon/cunits/v2"
	"github.com/hekmon/transmissionrpc"
)

// TransmissionTorrent is a generated, pointer-free, and less safe variant of transmissionrpc.Torrent to make using the 
// expr package easier.
type TransmissionTorrent struct {
	ActivityDate time.Time
	AddedDate time.Time
	BandwidthPriority int64
	Comment string
	CorruptEver int64
	Creator string
	DateCreated time.Time
	DesiredAvailable int64
	DoneDate time.Time
	DownloadDir string
	DownloadedEver int64
	DownloadLimit int64
	DownloadLimited bool
	Error int64
	ErrorString string
	Eta int64
	EtaIdle int64
	Files []*transmissionrpc.TorrentFile
	FileStats []*transmissionrpc.TorrentFileStat
	HashString string
	HaveUnchecked int64
	HaveValid int64
	HonorsSessionLimits bool
	ID int64
	IsFinished bool
	IsPrivate bool
	IsStalled bool
	LeftUntilDone int64
	MagnetLink string
	ManualAnnounceTime int64
	MaxConnectedPeers int64
	MetadataPercentComplete float64
	Name string
	PeerLimit int64
	Peers []*transmissionrpc.Peer
	PeersConnected int64
	PeersFrom transmissionrpc.TorrentPeersFrom
	PeersGettingFromUs int64
	PeersSendingToUs int64
	PercentDone float64
	Pieces string
	PieceCount int64
	PieceSize cunits.Bits
	Priorities []int64
	QueuePosition int64
	RateDownload int64
	RateUpload int64
	RecheckProgress float64
	SecondsDownloading int64
	SecondsSeeding time.Duration
	SeedIdleLimit int64
	SeedIdleMode int64
	SeedRatioLimit float64
	SeedRatioMode transmissionrpc.SeedRatioMode
	SizeWhenDone cunits.Bits
	StartDate time.Time
	Status transmissionrpc.TorrentStatus
	Trackers []*transmissionrpc.Tracker
	TrackerStats []*transmissionrpc.TrackerStats
	TotalSize cunits.Bits
	TorrentFile string
	UploadedEver int64
	UploadLimit int64
	UploadLimited bool
	UploadRatio float64
	Wanted []bool
	WebSeeds []string
	WebSeedsSendingToUs int64
}

// ToTransmissionTorrent converts the library struct to our generated struct.
func ToTransmissionTorrent(input transmissionrpc.Torrent) TransmissionTorrent {
	return TransmissionTorrent{
		ActivityDate: *input.ActivityDate,
		AddedDate: *input.AddedDate,
		BandwidthPriority: *input.BandwidthPriority,
		Comment: *input.Comment,
		CorruptEver: *input.CorruptEver,
		Creator: *input.Creator,
		DateCreated: *input.DateCreated,
		DesiredAvailable: *input.DesiredAvailable,
		DoneDate: *input.DoneDate,
		DownloadDir: *input.DownloadDir,
		DownloadedEver: *input.DownloadedEver,
		DownloadLimit: *input.DownloadLimit,
		DownloadLimited: *input.DownloadLimited,
		Error: *input.Error,
		ErrorString: *input.ErrorString,
		Eta: *input.Eta,
		EtaIdle: *input.EtaIdle,
		Files: input.Files,
		FileStats: input.FileStats,
		HashString: *input.HashString,
		HaveUnchecked: *input.HaveUnchecked,
		HaveValid: *input.HaveValid,
		HonorsSessionLimits: *input.HonorsSessionLimits,
		ID: *input.ID,
		IsFinished: *input.IsFinished,
		IsPrivate: *input.IsPrivate,
		IsStalled: *input.IsStalled,
		LeftUntilDone: *input.LeftUntilDone,
		MagnetLink: *input.MagnetLink,
		ManualAnnounceTime: *input.ManualAnnounceTime,
		MaxConnectedPeers: *input.MaxConnectedPeers,
		MetadataPercentComplete: *input.MetadataPercentComplete,
		Name: *input.Name,
		PeerLimit: *input.PeerLimit,
		Peers: input.Peers,
		PeersConnected: *input.PeersConnected,
		PeersFrom: *input.PeersFrom,
		PeersGettingFromUs: *input.PeersGettingFromUs,
		PeersSendingToUs: *input.PeersSendingToUs,
		PercentDone: *input.PercentDone,
		Pieces: *input.Pieces,
		PieceCount: *input.PieceCount,
		PieceSize: *input.PieceSize,
		Priorities: input.Priorities,
		QueuePosition: *input.QueuePosition,
		RateDownload: *input.RateDownload,
		RateUpload: *input.RateUpload,
		RecheckProgress: *input.RecheckProgress,
		SecondsDownloading: *input.SecondsDownloading,
		SecondsSeeding: *input.SecondsSeeding,
		SeedIdleLimit: *input.SeedIdleLimit,
		SeedIdleMode: *input.SeedIdleMode,
		SeedRatioLimit: *input.SeedRatioLimit,
		SeedRatioMode: *input.SeedRatioMode,
		SizeWhenDone: *input.SizeWhenDone,
		StartDate: *input.StartDate,
		Status: *input.Status,
		Trackers: input.Trackers,
		TrackerStats: input.TrackerStats,
		TotalSize: *input.TotalSize,
		TorrentFile: *input.TorrentFile,
		UploadedEver: *input.UploadedEver,
		UploadLimit: *input.UploadLimit,
		UploadLimited: *input.UploadLimited,
		UploadRatio: *input.UploadRatio,
		Wanted: input.Wanted,
		WebSeeds: input.WebSeeds,
		WebSeedsSendingToUs: *input.WebSeedsSendingToUs,
	}
}
