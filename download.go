package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type TorrentInfo struct {
	Name     string
	Stats    torrent.TorrentStats
	Seeding  bool
	Info     *metainfo.Info
	Spec     *torrent.TorrentSpec
	Progress float64
	Hash     metainfo.Hash
}

func init() {
	http.HandleFunc("/torrent/list", handleListDownloadTasks)
	http.HandleFunc("/torrent/addTorrent", handleTorrentAdd)
	http.HandleFunc("/torrent/addMagnet", handleMagnetAdd)
	http.HandleFunc("/torrent/startTorrent", startTorrentDownload)
	http.HandleFunc("/torrent/stopTorrent", stopTorrentDownload)
	http.HandleFunc("/torrent/dropTorrent", dropTorrent)
}

func NewTorrentClient() (*torrent.Client, error) {
	os.MkdirAll("./tmp", 0755)
	//Generate a Config for this client
	config := torrent.NewDefaultClientConfig()
	//Modify a few identifers
	config.HTTPUserAgent = "ArozOS TorrentA/1.0"
	config.ExtendedHandshakeClientVersion = "arozos.torrenta dev 20210218"
	config.UpnpID = "arozos/torrent"
	config.DataDir = "./download/"

	//Genrete new client and return it
	return torrent.NewClient(config)
}

//Start downloading a torrent file givne it hash
func startTorrentDownload(w http.ResponseWriter, r *http.Request) {
	hash, err := mv(r, "hash", true)
	if err != nil {
		sendErrorResponse(w, "Hash not defined")
		return
	}

	//Get the torrent client from request
	tc, err := getTorrentClientByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Get torrent by hash
	targetTorrent, ok := tc.Torrent(metainfo.NewHashFromHex(hash))
	if !ok {
		sendErrorResponse(w, "Torrent with hash not found")
		return
	}

	//Start download and upload
	targetTorrent.AllowDataDownload()
	targetTorrent.AllowDataUpload()

	sendOK(w)
}

//Stop the torrent download by giveing the torrent hash
func stopTorrentDownload(w http.ResponseWriter, r *http.Request) {
	hash, err := mv(r, "hash", true)
	if err != nil {
		sendErrorResponse(w, "Hash not defined")
		return
	}

	//Get the torrent client from request
	tc, err := getTorrentClientByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Get torrent by hash
	targetTorrent, ok := tc.Torrent(metainfo.NewHashFromHex(hash))
	if !ok {
		sendErrorResponse(w, "Torrent with hash not found")
		return
	}

	//Stop the download and upload
	targetTorrent.DisallowDataDownload()
	targetTorrent.DisallowDataUpload()

	sendOK(w)

}

//Drop (or delete) the torrent but keeping the file
func dropTorrent(w http.ResponseWriter, r *http.Request) {
	hash, err := mv(r, "hash", true)
	if err != nil {
		sendErrorResponse(w, "Hash not defined")
		return
	}

	//Get the torrent client from request
	tc, err := getTorrentClientByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Get torrent by hash
	targetTorrent, ok := tc.Torrent(metainfo.NewHashFromHex(hash))
	if !ok {
		sendErrorResponse(w, "Torrent with hash not found")
		return
	}

	//Drop the torrent
	targetTorrent.Drop()

	sendOK(w)

}

//Add a torrent by giving the torrent file virtual path
func handleTorrentAdd(w http.ResponseWriter, r *http.Request) {
	vpath, err := mv(r, "file", true)
	if err != nil {
		sendErrorResponse(w, "file not defined")
		return
	}

	//Resolve the vpath into realpath
	rpath, err := resolveVirtalPath(w, r, vpath)
	if err != nil {
		sendErrorResponse(w, "File not found")
		return
	}

	//Check if the file exists
	if !fileExists(rpath) {
		sendErrorResponse(w, "File not exists")
		return
	}

	//Get torrent client by username
	tc, err := getTorrentClientByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Ok. Add this to torrent download file
	torrent, err := tc.AddTorrentFromFile(rpath)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Wait until the torrent is loaded
	torrent.GotInfo()

	//Start task by default
	torrent.DownloadAll()
	log.Println("Torrent added: ", torrent.Name())

	sendOK(w)
}

func handleMagnetAdd(w http.ResponseWriter, r *http.Request) {

}

func handleListDownloadTasks(w http.ResponseWriter, r *http.Request) {
	//Get username and torrent client
	username := getUsername(w, r)
	torrentClient, err := getTorrentClientByUsername(username)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	if torrentClient == nil {
		sendErrorResponse(w, "No torrent client found for this user")
		return
	}

	//List all the running downloads
	allTorrents := torrentClient.Torrents()
	torrentInfos := []TorrentInfo{}

	for _, thisTorrent := range allTorrents {
		//Check if the torrent is ready or not by take its info field
		torrentInfo := thisTorrent.Info()
		metainfo := thisTorrent.Metainfo()
		torrentSpec := torrent.TorrentSpecFromMetaInfo(&metainfo)
		downloadedPiecesCounts := thisTorrent.Stats().PiecesDirtiedGood
		downloadCount := downloadedPiecesCounts.Int64()

		torrentInfos = append(torrentInfos, TorrentInfo{
			Name:     thisTorrent.Name(),
			Stats:    thisTorrent.Stats(),
			Seeding:  thisTorrent.Seeding(),
			Info:     torrentInfo,
			Spec:     torrentSpec,
			Progress: float64(downloadCount) / float64(thisTorrent.NumPieces()) * 100.0,
			Hash:     thisTorrent.InfoHash(),
		})

	}

	js, _ := json.Marshal(torrentInfos)
	sendJSONResponse(w, string(js))
}
