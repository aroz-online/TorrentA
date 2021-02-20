package main

import (
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/anacrolix/torrent"
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

	//send ok as ready
	sendOK(w)
}

func getTorrentClientByUsername(username string) (*torrent.Client, error) {
	val, ok := torrentClientLists.Load(username)
	if !ok {
		return nil, errors.New("No torrent client found for this user")
	}

	return val.(*torrent.Client), nil
}

func getTorrentClientByRequest(w http.ResponseWriter, r *http.Request) (*torrent.Client, error) {
	username := getUsername(w, r)
	return getTorrentClientByUsername(username)
}

func shutdownAllTorrentClients() {
	log.Println("Shutting down TorrentA torrent clients")
	torrentClientLists.Range(func(key, value interface{}) bool {
		//Shutdown the handler
		log.Println("Shutting down torrent client for " + key.(string))
		value.(*torrent.Client).Close()
		return true
	})
}

func init() {
	//Register the endpoints related to client isolation functions
	torrentClientLists = sync.Map{}

	//Each user request this endpoint once when enter the homepage
	http.HandleFunc("/init", initUserTorrentClientIfNotExists)
}
