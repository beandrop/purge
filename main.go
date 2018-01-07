package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/nlopes/slack"
)

func must(err error, msg string) {
	if err != nil {
		log.Println(fmt.Errorf("%s: %v", msg, err.Error()))
		os.Exit(1)
	}
}

func deleteMessage(tries int, sleep time.Duration, f func() error) error {
	var err error
	for i := 0; ; i++ {
		err = f()
		if err == nil {
			return nil
		}

		if i >= tries-1 {
			break
		}

		time.Sleep(sleep)
	}
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var token = flag.String("token", "", "https://api.slack.com/custom-integrations/legacy-tokens")
	var channelName = flag.String("name", "", "channel name")
	var latest = flag.String("latest", "", "timestamp")

	flag.Parse()

	api := slack.New(*token)
	channels, err := api.GetChannels(true)
	must(err, "channel list")
	var purgeID string
	for _, channel := range channels {
		if channel.Name == *channelName {
			purgeID = channel.ID
		}
	}

	latestTimestamp := *latest
	if latestTimestamp == "" {
		latestTimestamp = strconv.FormatInt(time.Now().AddDate(0, 0, -1).Unix(), 10)
		must(err, "latestTimestamp")
	}

	params := slack.NewHistoryParameters()
	params.Latest = latestTimestamp
	history, err := api.GetChannelHistory(purgeID, params)
	must(err, "channel history")

	for _, message := range history.Messages {
		err := deleteMessage(3, 2*time.Second, func() (err error) {
			for _, reply := range message.Replies {
				_ = deleteMessage(3, 2*time.Second, func() (err error) {
					_, _, err = api.DeleteMessage(purgeID, reply.Timestamp)
					return
				})
			}
			_, _, err = api.DeleteMessage(purgeID, message.Timestamp)
			return
		})
		log.Print(err)
	}
}
