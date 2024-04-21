# Summary 

Provide permission to [Thunderbird](https://www.thunderbird.net/en-GB/) to access my Gmail details. 

Use dev tools in Thunderbird to read the client_id and secret used to generate tokens. 

Enable smtp console debug to check how the protocol works and how the credentials are combined and sent to `smtp.gmail.com` 

Implement the protocol in `Go`

# Details 

## Accoutn setup 

Thunderbird will automatically download the IMAP and SMTP (SSL/TLS) settings for my Gmail account. 

![alt thunderbird](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_000_auto_setup_config.png)

## Authoriza app 

Thunderbird redirects me to `accounts.google.com` so I can get authenticated, and give it permission to send emails on my behalf. 

![alt authorize](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_002_authorize_thunder_app.png)

![alt permission](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_003_thunder_permission.png)

![alt verify](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_004_gmail_app_authorized.png)

## Thunderbird debug 

- Access dev tools on Thunderbird to see network requests and console logs 
![alt dev tools](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_005_thunder_open_dev_tools.png)

- Increase the log level for mailnews.smtp.loglevel (https://wiki.mozilla.org/MailNews:Logging)
![alt log level](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_006_thunder_increase_smtp_log_level.png)

- Send an email to capture an OAUTH2 authentication to Google. 
![alt token](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_007_dev_tools_capture_client_id_secret.png)

- Check the browser console for SMTP message exchange. Click on `SmtpClient.jsm` to look at the suppressed logs.
![alt console](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_008_console_smtp_protocol.png)

- Add a breaking point to read the value of the variable suppressed in the logs in the current scope. 
![alt scope](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_009_debugger_breakpoint_read_auth.png)

- Base64 decode the token, and print the byte value of the result to look at the separators (`\0x01`).
![alt decode](https://github.com/pzolo85/research/blob/f3407b3f0bbf6befcb6dc7ef7ac88bd0763a10be/images/gmail_010_decode_auth_format.png)

## Implement in Go 

```Go
[...]
// Create a TLS connection
	conn, err := tls.Dial("tcp", "smtp.gmail.com:465", &tls.Config{
		InsecureSkipVerify: true,
	})
// Create a new SMTP client passing the existing connection 
	cli, err := smtp.NewClient(conn, "smtp.gmail.com")
// Say Hi
	err = cli.Hello("localhost")
// Authenticate with our authenticator 
	err = cli.Auth(&GoogleAuth{})
[...]


// Authenticator methods 
func (ga *GoogleAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// OAUTH details
	host := "https://www.googleapis.com/oauth2/v3/token"

// Request a token with Thunderbird's client id and secret 
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
[...]

// Return the Credentials including user and token with x01 as separator 
return "XOAUTH2", []byte("user=" + cfg.Mail + "\x01" + "auth=Bearer " + google_token.Token + "\x01\x01"), nil
}
```


