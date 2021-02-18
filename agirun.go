package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
)

func getUsername(w http.ResponseWriter, r *http.Request) string {
	username, _ := handler.GetUserInfoFromRequest(w, r)
	return username
}

func runAGIContent(w http.ResponseWriter, r *http.Request, script string) (string, error) {
	//Get username and token from request
	_, token := handler.GetUserInfoFromRequest(w, r)

	//Execute the AGI request on server side
	resp, err := handler.RequestGatewayInterface(token, script)
	if err != nil {
		//Something went wrong when performing POST request
		log.Println(err)
	} else {
		//Try to read the resp body
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		resp.Body.Close()

		//Return the body content
		return string(bodyBytes), nil

	}

	return "", errors.New("Unknown error occured")
}
