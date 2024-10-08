package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/dunglas/mercure"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var key = uuid.NewString() // key used to sign / verify JWT

func fail(e error) {
	if e != nil {
		panic(e)
	}
}

type Msg struct {
	Msg  string `json:"msg"`
	Type string `json:"type"` //private, group, global
	To   string `json:"to"`   // <user> or <group>
}

func main() {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	mercureHub, err := mercure.NewHub(
		mercure.WithAnonymous(), // allow subscribers without JWT
		mercure.WithPublisherJWT([]byte(key), jwt.SigningMethodHS256.Name), // JWT key for publisher JWT
		mercure.WithDebug(),
	)
	fail(err)

	e.GET("/sub/messages", func(c echo.Context) error {
		user := c.QueryParam("user")
		group := c.QueryParam("group")
		w := c.Response().Writer
		r, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("mercure?topic=%s&topic=%s&topic=global", user, group),
			nil,
		)
		fail(err)

		r.Header.Add(echo.HeaderAccept, "text/event-stream")
		r = r.WithContext(c.Request().Context())

		mercureHub.SubscribeHandler(w, r)
		return nil
	})

	e.POST("/pub/messages", func(c echo.Context) error {
		var msg Msg
		err := c.Bind(&msg)
		fail(err)
		data := fmt.Sprintf("data=%s&id&type=%s&retry=&topic=%s", url.QueryEscape(msg.Msg), msg.Type, msg.To)
		if msg.Type == "global" {
			data = fmt.Sprintf("data=%s&id&type=global&retry=&topic=global", url.QueryEscape(msg.Msg))
		}

		bodyReader := strings.NewReader(data)

		// create a sign a mercure claim that allows to publish on any topic
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"mercure": map[string][]string{
				"publish": {"*"},
			},
		})
		token, err := t.SignedString([]byte(key))
		fail(err)

		w := c.Response().Writer
		r, err := http.NewRequest(
			http.MethodPost,
			"mercure",
			bodyReader,
		)
		fail(err)
		r.Header.Add(echo.HeaderContentType, echo.MIMEApplicationForm)
		r.Header.Add(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		r = r.WithContext(c.Request().Context())

		mercureHub.PublishHandler(w, r)
		return nil
	})

	e.Static("/", "./")
	e.Start(":9999")
}
