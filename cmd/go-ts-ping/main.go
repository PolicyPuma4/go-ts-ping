package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type user struct {
	id   int
	name string
}

func getUsers(client *http.Client) ([]user, error) {
	req, err := http.NewRequest("GET",
		fmt.Sprintf(
			"%s/%s/clientlist",
			os.Getenv("WEBQUERY_URL"),
			os.Getenv("WEBQUERY_SERVER_ID"),
		), nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", os.Getenv("WEBQUERY_API_KEY"))
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := struct {
		Body []struct {
			ClientDatabaseID string `json:"client_database_id"`
			ClientNickname   string `json:"client_nickname"`
		}

		Status struct {
			Code    int
			Message string
		}
	}{}

	if err := json.Unmarshal(bytes, &response); err != nil {
		return nil, err
	}

	if response.Status.Code != 0 {
		return nil, err
	}

	users := []user{}
	for _, client := range response.Body {
		id, err := strconv.Atoi(client.ClientDatabaseID)
		if err != nil {
			return nil, err
		}

		users = append(users, user{
			id:   id,
			name: client.ClientNickname,
		})
	}

	return users, nil
}

func main() {
	log.SetFlags(log.Flags() | log.Llongfile)
	firstLoop := true
	sleep, err := strconv.Atoi(os.Getenv("SLEEP"))
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	users := []user{}
	templates := []string{
		"%s joined the party.",
		"%s is here.",
		"Welcome, %s. We hope you brought pizza.",
		"A wild %s appeared.",
		"%s just landed.",
		"%s just slid into the server.",
		"%s just showed up!",
		"Welcome %s. Say hi!",
		"%s hopped into the server.",
		"Everyone welcome %s!",
		"Glad you're here, %s.",
		"Good to see you, %s.",
		"Yay you made it, %s!",
	}

	for {
		if !firstLoop {
			time.Sleep(time.Duration(sleep) * time.Second)
		}

		currentUsers, err := getUsers(client)
		if err != nil {
			log.Println(err)
			continue
		}

		newUsers := []user{}
		for _, currentUser := range currentUsers {
			contains := false
			for _, user := range users {
				if user.id == currentUser.id {
					contains = true
					break
				}
			}

			if !contains {
				newUsers = append(newUsers, currentUser)
			}
		}

		users = currentUsers
		if firstLoop {
			firstLoop = false
			continue
		}

		if len(newUsers) == 0 {
			continue
		}

		messages := []string{}
		for _, user := range newUsers {
			messages = append(messages, fmt.Sprintf(templates[rand.Intn(len(templates))], user.name))
		}

		buf, err := json.Marshal(map[string]string{"content": strings.Join(messages, "\n")})
		if err != nil {
			log.Println(err)
			continue
		}

		resp, err := http.Post(os.Getenv("DISCORD_WEBHOOK_URL"), "application/json", bytes.NewBuffer(buf))
		if err != nil {
			log.Println(err)
			continue
		}

		if err := resp.Body.Close(); err != nil {
			log.Println(err)
		}

		if resp.StatusCode != 204 {
			log.Println(resp.Status)
		}
	}
}
