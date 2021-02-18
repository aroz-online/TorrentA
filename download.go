package main

import (
	"github.com/anacrolix/torrent"
)

func NewTorrentClient() (*torrent.Client, error) {

	//Generate a Config for this client
	config := torrent.NewDefaultClientConfig()
	//Modify a few identifers
	config.HTTPUserAgent = "ArozOS TorrentA/1.0"
	config.ExtendedHandshakeClientVersion = "arozos.torrenta dev 20210218"
	config.UpnpID = "arozos/torrent"

	//Genrete new client and return it
	return torrent.NewClient(config)
}
