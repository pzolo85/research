# Summary 

[Mercure](https://mercure.rocks/) is a solution for real-time communications.    
It includes a hub that handles subscribers, connected using [EventSource](https://developer.mozilla.org/en-US/docs/Web/API/EventSource)    
and publishers, that can send events via a POST request with a JWT token.    

Mercure could be an overkill to setup in small apps, this is an example on how to use mercure as a hub while using a simple authentication metod and without JWT.   

## Client 

The web-app has a single EventSource object that subscribes to three different topics and it's configured to handle three different types of events. 
```
    const evs = new EventSource(`/sub/messages?user=${user}&group=${group}`)
    attachEvent(evs,"private",pm)
    attachEvent(evs,"group",grpm)
    attachEvent(evs,"global",glbm)

[...]

function attachEvent(es, t , node){
    es.addEventListener(t,e => {
        const m = document.createElement("li")
        m.textContent = `${e.timeStamp} | ${e.data}` 
        node.appendChild(m)
    })
}
```
## The Hub 

The hub is created allowing anonymouse subscribers. We can control private messages using `EventSource` type and topic.

```
	mercureHub, err := mercure.NewHub(
		mercure.WithAnonymous(), // allow subscribers without JWT
		mercure.WithPublisherJWT([]byte(key), jwt.SigningMethodHS256.Name), // JWT key for publisher JWT
		mercure.WithDebug(),
	)
```

## Server (subscribe)

When a subscription event is received, we create a new `http.Request` with the specific topics, extracted from the `EventSource` request.   
The request is then provided to the hub to handle, including the `http.ResponseWriter` 

```
		r, err := http.NewRequest(
			http.MethodGet,
			fmt.Sprintf("mercure?topic=%s&topic=%s&topic=global", user, group),
			nil,
		)
		fail(err)

		r.Header.Add(echo.HeaderAccept, "text/event-stream")
		r = r.WithContext(c.Request().Context())

		mercureHub.SubscribeHandler(w, r)
```

## Server (publish)

We start by creating a mercure claim that allows the publisher to publish on any topic.    
This uses the same `key` we passed to `mercure.NewHub`

```
	// create a sign a mercure claim that allows to publish on any topic
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"mercure": map[string][]string{
			"publish": {"*"},
		},
	})
	token, err := t.SignedString([]byte(key))
```

The token is used inside the handler to pass an authentication header to the hub 

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

## Publisher example 
```
$curl -H 'content-type:application/json' localhost:9999/pub/messages -d '{"type":"private","msg":"charlie, how are you?","to":"charlie"}'
$curl -H 'content-type:application/json' localhost:9999/pub/messages -d '{"type":"global","msg":"hello everyone"}'
$curl -H 'content-type:application/json' localhost:9999/pub/messages -d '{"type":"group","msg":"I love tennis","to":"tennis"}'
```

## Client result

![alt browser_out](https://raw.githubusercontent.com/pzolo85/research/refs/heads/main/images/mercure-app.png)