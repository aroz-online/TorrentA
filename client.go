package main

import (
	"net/http"
	"sync"
)

/*
	Client isolation layer

	This script isolate users between others so that
	each user has their own bt download list

*/

var torrentClientLists sync.Map //Use to store username -> torrent client

func initUserTorrentClientIfNotExists(w http.ResponseWriter, r *http.Request) {
	username := getUsername(w, r)
	if username == "" {
		sendErrorResponse(w, "Username undefined")
		return
	}
	_, ok := torrentClientLists.Load(username)
	if !ok {
		//Create a new client for this user
		ntc, err := NewTorrentClient()
		if err != nil {
			sendErrorResponse(w, err.Error())
		}

		torrentClientLists.Store(username, ntc)
	} else {
		//Already exists. OK
	}
	sendOK(w)
}

func init() {
	//Register the endpoints related to client isolation functions
	torrentClientLists = sync.Map{}

	//Each user request this endpoint once when enter the homepage
	http.HandleFunc("/init", initUserTorrentClientIfNotExists)

}
