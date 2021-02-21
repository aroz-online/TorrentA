package main

/*
	Client isolation layer

	This script isolate users between others so that
	each user has their own bt download list

*/

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/cenkalti/rain/torrent"
	"github.com/jinzhu/copier"
)

var SessionList map[string]*torrent.Session

func GetUserByUsername(username string) *torrent.Session {
	thisUser := SessionList[username]
	return thisUser
}

func initSessionForUser(w http.ResponseWriter, r *http.Request) error {
	username := getUsername(w, r)

	userDesktopAbs, err := resolveVirtalPath(w, r, "user:/Download")
	if err != nil {
		return err
	}

	if !fileExists(userDesktopAbs) {
		os.MkdirAll(userDesktopAbs, 0755)
	}

	//Generate a default config for this user
	userDefaultConfig := newCustomConfig(userDesktopAbs)

	//Create a session for this user
	sess, err := torrent.NewSession(userDefaultConfig)
	if err != nil {
		return err
	}

	SessionList[username] = sess

	return nil
}

func getUserSessionByRequest(w http.ResponseWriter, r *http.Request) (*torrent.Session, error) {
	username := getUsername(w, r)
	val, ok := SessionList[username]
	if !ok {
		initSessionForUser(w, r)
		nval, nok := SessionList[username]
		if !nok {
			return nil, errors.New("Failed to create new session for user")
		}
		return nval, nil
	}

	return val, nil
}

func newCustomConfig(downloadDir string) torrent.Config {
	thisConfig := torrent.Config{}
	copier.Copy(&thisConfig, &torrent.DefaultConfig)
	thisConfig.Database = "./torrent.db"
	thisConfig.DataDir = downloadDir
	thisConfig.DataDirIncludesTorrentID = false
	thisConfig.PrivatePeerIDPrefix = "-ArozOS-"
	thisConfig.PrivateExtensionHandshakeClientVersion = "TorrentA 1.110"

	return thisConfig
}

func closeAllSessions() {
	for k, v := range SessionList {
		log.Println("*TorrentA* Closing Client for " + k)
		v.Close()
	}
}

func init() {
	SessionList = map[string]*torrent.Session{}
}
