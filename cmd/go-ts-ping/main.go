package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go-ts-ping/internal/teamspeak"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	clientList = []teamspeak.Client{}
	firstLoop  = true
	templates  = []string{
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

	discordWebhookURL = os.Getenv("DISCORD_WEBHOOK_URL")
)

func loop() error {
	newClientList, err := teamspeak.GetClientList()
	if err != nil {
		return err
	}

	newClients := []teamspeak.Client{}
	for _, newClient := range newClientList {
		contains := false
		for _, user := range clientList {
			if user.ClientDatabaseID == newClient.ClientDatabaseID {
				contains = true
				break
			}
		}

		if !contains {
			newClients = append(newClients, newClient)
		}
	}

	clientList = newClientList
	if firstLoop {
		firstLoop = false
		return nil
	}

	if len(newClients) == 0 {
		return nil
	}

	messages := []string{}
	for _, client := range newClients {
		clientInfo, err := teamspeak.GetClientInfo(client.CLID)
		if err != nil {
			log.Println(err)
			continue
		}

		nickname := client.ClientNickname
		if clientInfo.ClientCountry != "" {
			nickname = fmt.Sprintf(
				"%s :flag_%s:",
				nickname,
				strings.ToLower(clientInfo.ClientCountry),
			)
		}

		messages = append(
			messages,
			fmt.Sprintf(templates[rand.Intn(len(templates))], nickname),
		)
	}

	buf, err := json.Marshal(map[string]string{"content": strings.Join(messages, "\n")})
	if err != nil {
		return err
	}

	resp, err := http.Post(discordWebhookURL, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}

	if err := resp.Body.Close(); err != nil {
		return err
	}

	if resp.StatusCode != 204 {
		return err
	}

	return nil
}

func main() {
	log.SetFlags(log.Flags() | log.Llongfile)
	sleep, err := strconv.Atoi(os.Getenv("SLEEP"))
	if err != nil {
		log.Fatal(err)
	}

	for ; true; <-time.Tick(time.Duration(sleep) * time.Second) {
		if err := loop(); err != nil {
			log.Println(err)
		}
	}
}
