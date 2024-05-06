package teamspeak

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var (
	queryURL      = os.Getenv("WEBQUERY_URL")
	queryServerID = os.Getenv("WEBQUERY_SERVER_ID")
	queryApiKey   = os.Getenv("WEBQUERY_API_KEY")
	client        = &http.Client{}
)

type Client struct {
	CLID             string `json:"clid"`
	ClientDatabaseID string `json:"client_database_id"`
	ClientNickname   string `json:"client_nickname"`
}

func GetClientList() ([]Client, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/%s/clientlist",
			queryURL,
			queryServerID,
		),
		nil,
	)

	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", queryApiKey)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := struct {
		Body   []Client `json:"body"`
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
	}{}

	if err := json.Unmarshal(bytes, &body); err != nil {
		return nil, err
	}

	if body.Status.Code != 0 {
		return nil, errors.New(body.Status.Message)
	}

	return body.Body, nil
}

type ClientInfo struct {
	ClientCountry string `json:"client_country"`
}

func GetClientInfo(clid string) (ClientInfo, error) {
	req, err := http.NewRequest(
		"GET",
		fmt.Sprintf(
			"%s/%s/clientinfo",
			queryURL,
			queryServerID,
		),
		nil,
	)

	if err != nil {
		return ClientInfo{}, err
	}

	query := req.URL.Query()
	query.Add("clid", clid)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("x-api-key", queryApiKey)
	resp, err := client.Do(req)
	if err != nil {
		return ClientInfo{}, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if resp.StatusCode != 200 {
		return ClientInfo{}, errors.New(resp.Status)
	}

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return ClientInfo{}, err
	}

	body := struct {
		Body   []ClientInfo `json:"body"`
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
	}{}

	if err := json.Unmarshal(bytes, &body); err != nil {
		return ClientInfo{}, err
	}

	if body.Status.Code != 0 {
		return ClientInfo{}, errors.New(body.Status.Message)
	}

	return ClientInfo{}, nil
}
