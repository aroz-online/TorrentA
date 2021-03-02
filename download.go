package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"

	"github.com/cenkalti/rain/torrent"
)

func init() {
	http.HandleFunc("/torrent/list", handleListDownloadTasks)
	http.HandleFunc("/torrent/addTorrent", handleTorrentAdd)
	http.HandleFunc("/torrent/addMagnet", handleMagnetAdd)
	http.HandleFunc("/torrent/startTorrent", startTorrentDownload)
	http.HandleFunc("/torrent/stopTorrent", stopTorrentDownload)
	http.HandleFunc("/torrent/dropTorrent", dropTorrent)
	http.HandleFunc("/torrent/startAll", handleStartAll)
	http.HandleFunc("/torrent/stopAll", handleStopAll)
}

type DownloadStatus struct {
	ID       string
	Name     string
	AddTime  int64
	Trackers []torrent.Tracker
	Stats    torrent.Stats
}

//Add a torrent by giving the torrent file virtual path
func handleTorrentAdd(w http.ResponseWriter, r *http.Request) {
	//Resolve the torrent file path
	vpath, err := mv(r, "file", true)
	if err != nil {
		sendErrorResponse(w, "Invalid file path")
		return
	}

	rpath, err := resolveVirtalPath(w, r, vpath)
	if err != nil {
		sendErrorResponse(w, "Invalid file path")
		return
	}

	if !fileExists(rpath) {
		sendErrorResponse(w, "Torrent file not found")
		return
	}

	//Create a session for this user if not exists
	sess, err := getUserSessionByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	torrentFileIOReader, err := os.Open(rpath)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}
	sess.AddTorrent(torrentFileIOReader, nil)

	sendOK(w)
}

func handleMagnetAdd(w http.ResponseWriter, r *http.Request) {
	//Resolve the torrent file path
	magnet, err := mv(r, "magnet", true)
	if err != nil {
		sendErrorResponse(w, "Invalid file path")
		return
	}

	//Create a session for this user if not exists
	sess, err := getUserSessionByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	_, err = sess.AddURI(magnet, nil)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}
	sendOK(w)
}

//Start downloading a torrent file givne it hash
func startTorrentDownload(w http.ResponseWriter, r *http.Request) {
	id, err := mv(r, "hash", true)
	if err != nil {
		sendErrorResponse(w, "Invalid torrent hash")
		return
	}

	//Get user session
	sess, err := getUserSessionByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Start the torrent download
	targetTorrent := sess.GetTorrent(id)
	if targetTorrent == nil {
		sendErrorResponse(w, "Target torrent not found")
		return
	}

	targetTorrent.Start()

	sendOK(w)
}

//Stop the torrent download by giveing the torrent hash
func stopTorrentDownload(w http.ResponseWriter, r *http.Request) {
	id, err := mv(r, "hash", true)
	if err != nil {
		sendErrorResponse(w, "Invalid torrent hash")
		return
	}

	//Get user session
	sess, err := getUserSessionByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//Stop the target torrent
	targetTorrent := sess.GetTorrent(id)
	if targetTorrent == nil {
		sendErrorResponse(w, "Target torrent not found")
		return
	}

	targetTorrent.Stop()

	sendOK(w)

}

//Drop (or delete) the torrent but keeping the file
func dropTorrent(w http.ResponseWriter, r *http.Request) {
	id, err := mv(r, "hash", true)
	if err != nil {
		sendErrorResponse(w, "Invalid torrent hash")
		return
	}

	//Get user session
	sess, err := getUserSessionByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	//delete the torrent
	err = sess.RemoveTorrent(id)
	if err != nil {
		sendErrorResponse(w, "Unable to remove torrent with given id")
		return
	}

	sendOK(w)

}

//List all the dowloading tasks for this user
func handleListDownloadTasks(w http.ResponseWriter, r *http.Request) {
	sess, err := getUserSessionByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	torrentList := sess.ListTorrents()
	results := []DownloadStatus{}
	for _, thisTorrent := range torrentList {
		results = append(results, DownloadStatus{
			thisTorrent.ID(),
			thisTorrent.Name(),
			thisTorrent.AddedAt().Unix(),
			thisTorrent.Trackers(),
			thisTorrent.Stats(),
		})
	}

	//Sort the list
	sort.Slice(results, func(i, j int) bool {
		return results[i].AddTime < results[j].AddTime
	})

	js, _ := json.Marshal(results)
	sendJSONResponse(w, string(js))
}

func handleStartAll(w http.ResponseWriter, r *http.Request) {
	sess, err := getUserSessionByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	sess.StartAll()

	sendOK(w)
}

func handleStopAll(w http.ResponseWriter, r *http.Request) {
	sess, err := getUserSessionByRequest(w, r)
	if err != nil {
		sendErrorResponse(w, err.Error())
		return
	}

	sess.StopAll()

	sendOK(w)
}
