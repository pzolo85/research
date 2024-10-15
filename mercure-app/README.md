# Summary

[Mercure](https://mercure.rocks/) is a solution for real-time communications.
It includes a hub that handles subscribers, connected using [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)
and publishers, that can send events via a POST request with a JWT token.

Mercure could be an overkill to setup in small apps, this is an example on how to use mercure as a hub while using a simple authentication metod and without JWT.

# Web Client

The `app/app.mjs` has a single EventSource object that uses query parameters to specify the user, group and appID.
This parameters are parsed by the backend and converted into `mercure` topics.
Each topic represents the name space where this `EventSource` will receive events.
These are the topics that the backend creates from the query string:
- user: private messages for this user.
- group: messages for members of this group.
- global: messages for all users.
- appID: each time the app starts, it generates a random appID. This topic is used to send cached private messages for the user.
```
  // Create eventsource specifying user, group and instance ID
  const appID = self.crypto.randomUUID();
  const evs = new EventSource(
    `/sub/messages?user=${user}&group=${group}&appid=${appID}`
  );
```
The `EventSource` has a default `onmessage` handler to handle messages that don't have a defined `type`.
We're going to use different handlers for each type of event (`private`, `group`, `global`) to attach the messages to a different `<ul>`
```
  attachEvent(evs, "private", pm);
  attachEvent(evs, "group", grpm);
  attachEvent(evs, "global", glbm);

  evs.onmessage = (event) => {
    console.log("received unknown event type");
    console.log(event);
  };
[...]
function attachEvent(es, t, node) {
  es.addEventListener(t, (e) => {
    const m = document.createElement("li");
    m.textContent = `${e.timeStamp} | ${e.data}`;
    node.appendChild(m);
  });
}
```
![alt browser_out](https://raw.githubusercontent.com/pzolo85/research/refs/heads/main/images/mercure-app.png)

# Backend application

## The Hub

The hub is created allowing anonymous subscribers.
We can which user receives a message by setting a specific `topic` when we publish a message.
- topic set to `alice`: messages are routed only to that user
- topic set to `interfaces`: messages are routed to all users that belong to the group `interfaces`
- topic set to `global`: messages are routed to all users
- topic set to `<appid>`: messages are routed to a specific instance of an app (this is used to sned private cached messages)
```
	mercureHub, err = mercure.NewHub(
		mercure.WithAnonymous(), // allow subscribers without JWT
		mercure.WithPublisherJWT([]byte(key), jwt.SigningMethodHS256.Name), // key used to validate signature of JWT
		mercure.WithDebug(),
	)
```

## Server (subscribe)

When a subscription event is received, we create a new `http.Request` with the specific topics, extracted from the `EventSource` request.
The request is then sent to the `mercure.Hub` to handle, including the `http.ResponseWriter`
```
	r, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("mercure?topic=%s&topic=%s&topic=%s&topic=global", user, group, appID),
		nil,
	)
[...]
	r.Header.Add(echo.HeaderAccept, "text/event-stream")
	r = r.WithContext(c.Request().Context())

	// launch a new routine to publish all cached messages for <user> in the <appID> topic
	go replayPM(user, appID)
	mercureHub.SubscribeHandler(w, r)
```
Just before sending the request to the `mercure.Hub`, we launch a go routine that publishes cached messages
```
	for _, m := range msgs.([]string) {
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
```
The events are sent to the `<appID>` topic. If the user already had anoter session opened in a diferent browser, that session won't receive the cached private messages.
Notice how we use an `httptest.NewRecorder()` to create a struct that complies with `http.ResponseWriter`

## Server (publish)

We start by creating a mercure claim that allows the publisher to publish on any topic.
This uses the same `key` we passed to `mercure.NewHub`

```
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"mercure": map[string][]string{
			"publish": {"*"},
		},
	})
	token, err := t.SignedString([]byte(key))
```
Private messages are cached in memory for 24h to be delivered to the users even if they're not subscribed when the event is published
```
	if msg.Type == "private" {
		usrmsgs, ok := msgCache.Get(msg.To)
		if !ok {
			usrmsgs = make([]string, 0)
		}
		msgs := usrmsgs.([]string)
		msgs = append(msgs, msg.Msg)
		msgCache.Set(msg.To, msgs, cache.DefaultExpiration)
	}
```
The JWT token is used inside the handler to pass an authentication header to the hub
```
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
		r.Header.Add(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
		r = r.WithContext(c.Request().Context())
		mercureHub.PublishHandler(w, r)
```
The response is sent back to the client that tires to publish the data via the `http.ResponseWriter`

# Golang Client

There's a simple implementation of a golang client on `client/main.go`.
The client reads the `user` and `group` from cli flags to create a subscription to the hub
```
	r, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("http://localhost:9999/sub/messages?user=%s&group=%s&appid=%s", user, group, appID),
		nil)
	if err != nil {
		panic(err)
	}
```
A new random `appID` is generated every time the client is launched.
The  client creates a `*bufio.Scanner` to read the `*http.Response.Body` line by line.
```
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
```

# Example
- Launch the hub from `./app`
```
$ go run main.go
```
- Create a new client from `./client`
```
$ go run main.go -u bob -g baseball
```
- The hub reports that a new client connected, and lists all the topics it subscribes to
```
2024-10-09T14:07:01.109+0100    INFO    mercure@v0.16.3/subscribe.go:205        New subscriber  {"subscriber": {"id": "urn:uuid:fdafcc02-3f65-4d36-acf2-11e433bace8c", "last_event_id": "", "topics": ["bob", "baseball", "ae8d5bc1-1aa0-4a99-a571-ba90b0fa8622", "global"]}}
```
- Publish a new message for group "baseball"
```
$ curl -H 'content-type:application/json' localhost:9999/pub/messages -d '{"msg":"hi from curl","type":"group","to":"baseball"}'
urn:uuid:7a1e75fc-3811-4b29-8a1d-d671b5cf40bb
```
- The hub reports it received a message
```
2024-10-09T14:10:46.031+0100    DEBUG   mercure@v0.16.3/publish.go:79   Update published        {"update": {"id": "urn:uuid:7a1e75fc-3811-4b29-8a1d-d671b5cf40bb", "type": "group", "retry": 0, "topics": ["baseball"], "private": false, "data": "hi from curl"}, "remote_addr": ""}
2024-10-09T14:10:46.031+0100    DEBUG   mercure@v0.16.3/subscribe.go:150        Update sent     {"subscriber": {"id": "urn:uuid:fdafcc02-3f65-4d36-acf2-11e433bace8c", "last_event_id": "", "topics": ["bob", "baseball", "ae8d5bc1-1aa0-4a99-a571-ba90b0fa8622", "global"]}, "update": {"id": "urn:uuid:7a1e75fc-3811-4b29-8a1d-d671b5cf40bb", "type": "group", "retry": 0, "topics": ["baseball"], "private": false, "data": "hi from curl"}}
```
- The client shows the message
```
$ go run main.go -u bob -g baseball
group-baseball | hi from curl
```
- Send a private message to bob
```
$ curl -H 'content-type:application/json' localhost:9999/pub/messages -d '{"msg":"hi bob","type":"private","to":"bob"}'
```
- The client shows the message
```
$ go run main.go -u bob -g baseball
group-baseball | hi from curl
private-bob | hi bob
```
- Start a new client in a different cli
```
$ go run main.go -u bob -g baseball
private-bob | cached msg >> hi bob
```
-  Cached message only gets delivered to the new instance.
- Send a global message
```
$ curl -H 'content-type:application/json' localhost:9999/pub/messages -d '{"msg":"bye all","type":"global"}'
```
- Hub reports delivering the message to multiple subscribers
```
2024-10-09T14:16:25.080+0100    DEBUG   mercure@v0.16.3/publish.go:79   Update published        {"update": {"id": "urn:uuid:42d6fc78-0f93-40d6-9cbf-b06fb316b9a5", "type": "global", "retry": 0, "topics": ["global"], "private": false, "data": "bye all"}, "remote_addr": ""}
2024-10-09T14:16:25.080+0100    DEBUG   mercure@v0.16.3/subscribe.go:150        Update sent     {"subscriber": {"id": "urn:uuid:fdafcc02-3f65-4d36-acf2-11e433bace8c", [...]
2024-10-09T14:16:25.080+0100    DEBUG   mercure@v0.16.3/subscribe.go:150        Update sent     {"subscriber": {"id": "urn:uuid:eda3cc23-33d9-49d1-83f5-11559ba6886f", [...]
```


