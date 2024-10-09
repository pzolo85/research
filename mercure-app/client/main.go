package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var (
	user  string
	group string
	appID string
)

const (
	event = "event: "
	data  = "data: "
)

func main() {
	appID = uuid.NewString()

	flag.StringVar(&user, "u", "", "user name")
	flag.StringVar(&group, "g", "", "group")
	flag.Parse()

	r, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("http://localhost:9999/sub/messages?user=%s&group=%s&appid=%s", user, group, appID),
		nil)
	if err != nil {
		panic(err)
	}

	res, err := http.DefaultClient.Do(r)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	var curEvent string
	var curData string

	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, event) {
			curEvent = text[len(event):]
		}
		if strings.HasPrefix(text, data) {
			curData = text[len(data):]
			printMsg(curEvent, curData)
			curEvent = ""
			curData = ""
		}
	}
}

func printMsg(event, data string) {
	switch event {
	case "global":
		fmt.Printf("%s | %s\n", event, data)
	case "group":
		fmt.Printf("%s | %s\n", event+"-"+group, data)
	case "private":
		fmt.Printf("%s | %s\n", event+"-"+user, data)
	default:
		fmt.Printf("unknown event | %s\n", data)
	}
}
