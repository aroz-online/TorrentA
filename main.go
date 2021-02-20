package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"arozos.com/TorrentA/mod/aroz"
	"github.com/anacrolix/torrent"
)

var (
	handler       *aroz.ArozHandler
	torrentClient *torrent.Client
)

/*
	TorrentA
	The torrent download fro ArozOS project
*/

//Kill signal handler. Do something before the system the core terminate.
func SetupCloseHandler() {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Println("\r- Shutting down TorrentA module.")
		//Do other things like close database or opened files

		//Shutdown alld torrent clients
		shutdownAllTorrentClients()

		os.Exit(0)
	}()
}

func main() {
	//Register the module
	handler = aroz.HandleFlagParse(aroz.ServiceInfo{
		Name:         "TorrentA",
		Desc:         "The torrent downloader for ArozOS",
		Group:        "Download",
		IconPath:     "TorrentA/img/small_icon.png",
		Version:      "0.1",
		StartDir:     "TorrentA/index.html",
		SupportFW:    true,
		LaunchFWDir:  "TorrentA/index.html",
		SupportEmb:   true,
		LaunchEmb:    "TorrentA/index.html",
		InitFWSize:   []int{1150, 640},
		InitEmbSize:  []int{1150, 640},
		SupportedExt: []string{".torrent"},
	})

	//Register the standard web services urls
	fs := http.FileServer(http.Dir("./web"))

	http.Handle("/", fs)

	//Setup close handler
	SetupCloseHandler()

	//Any log println will be shown in the core system via STDOUT redirection. But not STDIN.
	log.Println("*TorrentA* Started. Listening on " + handler.Port)
	err := http.ListenAndServe(handler.Port, nil)
	if err != nil {
		log.Fatal(err)
	}

}
