package gmail

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
)

type GoogleAuth struct{}

type GoogleConfig struct {
	Mail         string `json:"mail"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RefreshToken string `json:"refresh_token"`
}

var cfg *GoogleConfig = &GoogleConfig{}

func Send() {

	b, err := os.ReadFile("gmail/config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, cfg)
	if err != nil {
		panic(err)
	}

	conn, err := tls.Dial("tcp", "smtp.gmail.com:465", &tls.Config{
		InsecureSkipVerify: true,
	})

	if err != nil {
		panic(err)
	}

	cli, err := smtp.NewClient(conn, "smtp.gmail.com")
	if err != nil {
		panic(err)
	}
	err = cli.Hello("localhost")
	if err != nil {
		panic(err)
	}
	err = cli.Noop()
	if err != nil {
		panic(err)
	}

	err = cli.Auth(&GoogleAuth{})
	if err != nil {
		panic(err)
	}
	err = cli.Mail(cfg.Mail)
	if err != nil {
		panic(err)
	}
	err = cli.Rcpt("aleksei_karamazov@outlook.com")
	if err != nil {
		panic(err)
	}
	w, err := cli.Data()
	if err != nil {
		panic(err)
	}
	msg := []byte("To: aleksei_karamazov@outlook.com\r\n" +

		"Subject: discount Gophers!\r\n" +

		"\r\n" +

		"This is the email body.\r\n")
	n, err := w.Write(msg)
	if err != nil {
		panic(err)
	}
	fmt.Println("wrote ", n)
	err = w.Close()
	if err != nil {
		panic(err)
	}
	cli.Close()
}

func (ga *GoogleAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// OAUTH details
	host := "https://www.googleapis.com/oauth2/v3/token"

	resp, err := http.PostForm(host, url.Values{
		"client_id": {
			cfg.ClientID,
		},
		"client_secret": {
			cfg.ClientSecret,
		},
		"grant_type": {
			"refresh_token",
		},
		"refresh_token": {
			cfg.RefreshToken,
		},
	})

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var google_token struct {
		Token string `json:"access_token"`
	}
	fmt.Println(string(body))
	json.Unmarshal(body, &google_token)

	return "XOAUTH2", []byte("user=" + cfg.Mail + "\x01" + "auth=Bearer " + google_token.Token + "\x01\x01"), nil
}

func (ga *GoogleAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	fmt.Println("Next")
	fmt.Println("from server:", string(fromServer))
	fmt.Println("More? ", more)
	if more {
		return nil, fmt.Errorf("server wants whaaat??")
	}
	return nil, nil
}
