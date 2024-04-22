package outlook

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

type OutlookAuth struct{}

type OutlookConfig struct {
	Mail         string `json:"mail"`
	ClientID     string `json:"client_id"`
	RefreshToken string `json:"refresh_token"`
}

var cfg *OutlookConfig = &OutlookConfig{}

func Send() {

	b, err := os.ReadFile("outlook/config.json")
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(b, cfg)
	if err != nil {
		panic(err)
	}
	cli, err := smtp.Dial("smtp-mail.outlook.com:587")
	if err != nil {
		panic(err)
	}
	err = cli.StartTLS(&tls.Config{InsecureSkipVerify: true})
	if err != nil {
		panic(err)
	}
	err = cli.Auth(&OutlookAuth{})
	if err != nil {
		panic(err)
	}
	err = cli.Mail(cfg.Mail)
	if err != nil {
		panic(err)
	}
	err = cli.Rcpt("garciaokellydavism@gmail.com")
	if err != nil {
		panic(err)
	}
	w, err := cli.Data()
	if err != nil {
		panic(err)
	}
	msg := []byte("To: garciaokellydavism@gmail.com\r\n" +

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

func (ga *OutlookAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// OAUTH details
	host := "https://login.microsoftonline.com/common/oauth2/v2.0/token"

	resp, err := http.PostForm(host, url.Values{
		"client_id": {
			cfg.ClientID,
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
	var Outlook_token struct {
		Token string `json:"access_token"`
	}
	fmt.Println(string(body))
	json.Unmarshal(body, &Outlook_token)

	return "XOAUTH2", []byte("user=" + cfg.Mail + "\x01" + "auth=Bearer " + Outlook_token.Token + "\x01\x01"), nil
}

func (ga *OutlookAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	fmt.Println("Next")
	fmt.Println("from server:", string(fromServer))
	fmt.Println("More? ", more)
	if more {
		return nil, fmt.Errorf("server wants whaaat??")
	}
	return nil, nil
}
