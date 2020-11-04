package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
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

func requestGameInfo(id string) (*BGGResponse, error) {
	resp, err := http.Get("https://www.boardgamegeek.com/xmlapi2/thing?id=" + id)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data BGGResponse
	err = xml.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

type GameNotFound struct{}

func (e *GameNotFound) Error() string {
	return "Game not found"
}

type GameNameMissing struct{}

func (e *GameNameMissing) Error() string {
	return "Missing game name"
}

func extractGameName(bggResponse *BGGResponse) (string, error) {
	if len(bggResponse.BoardGames) == 0 {
		return "", &GameNotFound{}
	}

	for _, name := range bggResponse.BoardGames[0].Names {
		if name.Type == "primary" {
			return name.Name, nil
		}
	}

	return "", &GameNameMissing{}
}

type NameOrError struct {
	GameId string
	Name   string
	Error  error
}

func getGameName(gameId string, results chan<- NameOrError, wg *sync.WaitGroup) {
	var outcome NameOrError

	defer func() {
		results <- outcome
		wg.Done()
	}()

	data, err := requestGameInfo(gameId)
	if err != nil {
		outcome = NameOrError{GameId: gameId, Error: err}
		return
	}

	name, err := extractGameName(data)
	if err != nil {
		outcome = NameOrError{GameId: gameId, Error: err}
		return
	}

	outcome = NameOrError{GameId: gameId, Name: name}
}

func main() {
	var wg sync.WaitGroup
	ch := make(chan NameOrError)

	for i := 0; i < 100; i += 1 {
		wg.Add(1)
		go getGameName(fmt.Sprint(i), ch, &wg)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	outcomes := make(map[string]string)
	for gameInfo := range ch {
		if gameInfo.Error != nil {
			outcomes[gameInfo.GameId] = fmt.Sprint(gameInfo.Error)
		} else {
			outcomes[gameInfo.GameId] = gameInfo.Name
		}
	}

	fmt.Println(outcomes)
}
