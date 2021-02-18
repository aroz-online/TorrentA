package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"arozos.com/TorrentA/mod/aroz"
)

var (
	handler *aroz.ArozHandler
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

		os.Exit(0)
	}()
}

func main() {

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

	//To receive kill signal from the System core, you can setup a close handler to catch the kill signal
	//This is not nessary if you have no opened files / database running
	SetupCloseHandler()

	//Any log println will be shown in the core system via STDOUT redirection. But not STDIN.
	log.Println("TorrentA Started. Listening on " + handler.Port)
	err := http.ListenAndServe(handler.Port, nil)
	if err != nil {
		log.Fatal(err)
	}

}

//API Test Demo. This showcase how can you access arozos resources with RESTFUL API CALL
func apiTestDemo(w http.ResponseWriter, r *http.Request) {
	//Get username and token from request
	username, token := handler.GetUserInfoFromRequest(w, r)
	log.Println("Received request from: ", username, " with token: ", token)

	//Create an AGI Call that get the user desktop files
	script := `
		if (requirelib("filelib")){
			var filelist = filelib.glob("user:/Desktop/*")
			sendJSONResp(JSON.stringify(filelist));
		}else{
			sendJSONResp(JSON.stringify({
				error: "Filelib require failed"
			}));
		}
	`

	//Execute the AGI request on server side
	resp, err := handler.RequestGatewayInterface(token, script)
	if err != nil {
		//Something went wrong when performing POST request
		log.Println(err)
	} else {
		//Try to read the resp body
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			w.Write([]byte(err.Error()))
			return
		}
		resp.Body.Close()

		//Relay the information to the request using json header
		//Or you can process the information within the go program
		w.Header().Set("Content-Type", "application/json")
		w.Write(bodyBytes)

	}
}
