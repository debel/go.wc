package main

import (
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
)

type BoardGameName struct {
	Type string `xml:"type,attr"`
	Name string `xml:"value,attr"`
}

type BoardGame struct {
	XMLName   xml.Name        `xml:"item"`
	Thumbnail string          `xml:"thumbnail"`
	Names     []BoardGameName `xml:"name"`
}

type BGGResponse struct {
	XMLName    xml.Name    `xml:"items"`
	BoardGames []BoardGame `xml:"item"`
}

func parseXML(xmlString []byte) (BGGResponse, error) {
	var data BGGResponse
	err := xml.Unmarshal(xmlString, &data)
	return data, err
}

func requestGameInfo(id string) ([]byte, error) {
	resp, err := http.Get("https://www.boardgamegeek.com/xmlapi2/thing?id=" + id)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func main() {
	gameId := "1234567"
	response, err := requestGameInfo(gameId)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	data, err := parseXML(response)
	if err != nil {
		log.Fatalln(err)
		panic(err)
	}

	if len(data.BoardGames) == 0 {
		log.Fatalln("No game found with id " + gameId)
		panic(1)
	}

	for _, name := range data.BoardGames[0].Names {
		if name.Type == "primary" {
			log.Printf(name.Name)
		}
	}
}
