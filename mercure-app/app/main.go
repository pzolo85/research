package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"github.com/dunglas/mercure"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
)

var (
	key        = uuid.NewString() // key used to sign / verify JWT
	mercureHub *mercure.Hub
	msgCache   *cache.Cache // cache of private messages
	token      string       // JWT with permission to publish to all topics
)

type Msg struct {
	Msg  string `json:"msg"`
	Type string `json:"type"` //private, group, global
	To   string `json:"to"`   // <user> or <group>
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	var err error

	// create and sign a mercure claim that allows to publish on any topic
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"mercure": map[string][]string{
			"publish": {"*"},
		},
	})
	token, err = t.SignedString([]byte(key))
	fail(err)

	mercureHub, err = mercure.NewHub(
		mercure.WithAnonymous(), // allow subscribers without JWT
		mercure.WithPublisherJWT([]byte(key), jwt.SigningMethodHS256.Name), // key used to validate signature of JWT
		mercure.WithDebug(),
	)
	fail(err)

	// private messages are cached for 24 hours
	msgCache = cache.New(time.Hour*24, time.Hour)

	e.GET("/sub/messages", sub)
	e.POST("/pub/messages", pub)
	e.Static("/", "./")
	e.Start(":9999")
}

func sub(c echo.Context) error {
	user := c.QueryParam("user")
	group := c.QueryParam("group")
	appID := c.QueryParam("appid")

	// we suscribe the user to 4 topics:
	// global for global messgest
	// <group> for group messages, ex: baseball
	// <user> for private messages
	// <appID> this represents an unique instance of the user, it's used to reply cached messages
	w := c.Response().Writer
	r, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("mercure?topic=%s&topic=%s&topic=%s&topic=global", user, group, appID),
		nil,
	)
	fail(err)
	r.Header.Add(echo.HeaderAccept, "text/event-stream")
	r = r.WithContext(c.Request().Context())

	// launch a new routine to publish all cached messages for <user> in the <appID> topic
	go replayPM(user, appID)
	mercureHub.SubscribeHandler(w, r)
	return nil
}

func pub(c echo.Context) error {
	var msg Msg
	err := c.Bind(&msg)
	fail(err)
	if msg.Type == "global" {
		msg.To = "global"
	}

	// private messages are stored in the mem cache
	if msg.Type == "private" {
		usrmsgs, ok := msgCache.Get(msg.To)
		if !ok {
			usrmsgs = make([]string, 0)
		}
		msgs := usrmsgs.([]string)
		msgs = append(msgs, msg.Msg)
		msgCache.Set(msg.To, msgs, cache.DefaultExpiration)
	}

	// We publish messages of type global, private or group
	// The type is used in the front-end by the EventSource to selet the correct event handler
	// The topic is used by the `mercure.Hub` to select which subscribers are going to receive the event
	data := fmt.Sprintf("data=%s&id&type=%s&retry=&topic=%s", url.QueryEscape(msg.Msg), msg.Type, msg.To)
	bodyReader := strings.NewReader(data)

	w := c.Response().Writer
	r, err := http.NewRequest(
		http.MethodPost,
		"mercure",
		bodyReader,
	)
	fail(err)
	r.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
	// we add the JWT that allows publishing to ALL topics
	r.Header.Add(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
	r = r.WithContext(c.Request().Context())

	mercureHub.PublishHandler(w, r)
	return nil
}

func replayPM(user, appID string) {
	// Extract all cached messages for this user
	msgs, ok := msgCache.Get(user)
	if !ok {
		return
	}

	for _, m := range msgs.([]string) {
		// We send the cached messages
		// type is set to "private", so the EventSource in the front end places the message in the private section
		// The topic is set to the specific app instance, so existing sessions of the same user don't get cached messages
		data := fmt.Sprintf("data=%s&id&type=%s&retry=&topic=%s", url.QueryEscape("cached msg >> "+m), "private", appID)
		w := httptest.NewRecorder()
		bodyReader := strings.NewReader(data)
		r, err := http.NewRequest(
			http.MethodPost,
			"mercure",
			bodyReader,
		)
		fail(err)
		r.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
		r.Header.Add(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))

		mercureHub.PublishHandler(w, r)
	}
}

func fail(e error) {
	if e != nil {
		panic(e)
	}
}
