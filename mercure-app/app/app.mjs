document.body.onload = start;

function start() {
  window.setTimeout(generate, 5 * 1000);
  // html stuff
  const pm = document.getElementById("msg");
  const grpm = document.getElementById("grpmsg");
  const glbm = document.getElementById("glbmsg");
  const ui = Math.floor(Math.random() * users.length);
  const gi = Math.floor(Math.random() * groups.length);
  const user = users[ui];
  const group = groups[gi];
  const t = document.getElementById("title");
  t.textContent = `Hello ${user}`;
  const gt = document.getElementById("grptitle");
  gt.textContent = `Group ${group} messages:`;

  // Create eventsource specifying user, group and instance ID
  const appID = self.crypto.randomUUID();
  const evs = new EventSource(
    `/sub/messages?user=${user}&group=${group}&appid=${appID}`
  );
  attachEvent(evs, "private", pm);
  attachEvent(evs, "group", grpm);
  attachEvent(evs, "global", glbm);

  evs.onmessage = (event) => {
    console.log("received unknown event type");
    console.log(event);
  };

  evs.onerror = (event) => {
    console.log("received an error");
    console.log(event);
  };
}

function attachEvent(es, t, node) {
  es.addEventListener(t, (e) => {
    const m = document.createElement("li");
    m.textContent = `${e.timeStamp} | ${e.data}`;
    node.appendChild(m);
  });
}

function generate() {
  const mi = Math.floor(Math.random() * messages.length);
  const ui = Math.floor(Math.random() * users.length);
  const gi = Math.floor(Math.random() * groups.length);
  const ti = Math.floor(Math.random() * types.length);

  const user = users[ui];
  const group = groups[gi];
  const type = types[ti];
  const msg = messages[mi];

  let reqBody;
  switch (type) {
    case "private":
      reqBody = { type, msg, to: user };
      console.log(`sending message of type ${type} to ${user}`);
      break;
    case "group":
      reqBody = { type, msg, to: group };
      console.log(`sending message of type ${type} to ${group}`);
      break;
    default:
      reqBody = { type, msg };
      console.log(`sending message of type ${type}`);
      break;
  }

  const r = new Request("/pub/messages", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(reqBody),
  });
  fetch(r);

  setTimeout(generate, 5 * 1000);
}

const types = ["private", "group", "global"];
const groups = ["interfaces", "types", "functions"];
const users = ["alice", "bob", "charlie", "david", "eve", "frank"];
const messages = [
  "Go is known for its simplicity.",
  "Concurrency is a key feature of Go.",
  "Go was created at Google.",
  "Go compiles quickly.",
  "Go has garbage collection.",
  "Go supports static typing.",
  "Go is often used for web development.",
  "Go has a strong standard library.",
  "Go is open source.",
  "Go uses goroutines for concurrency.",
  "Go's syntax is clean and concise.",
  "Go is great for building microservices.",
  "Go has built-in testing tools.",
  "Go supports cross-compilation.",
  "Go is designed for scalability.",
  "Go has a growing community.",
  "Go's package management is straightforward.",
  "Go is used by many large companies.",
  "Go's error handling is explicit.",
  "Go has a simple dependency management system.",
  "Go's performance is impressive.",
  "Go supports interfaces for polymorphism.",
  "Go's tooling is robust.",
  "Go has a unique approach to object-oriented programming.",
  "Go's documentation is comprehensive.",
  "Go's standard library includes a web server.",
  "Go's concurrency model is based on CSP.",
  "Go's memory management is efficient.",
  "Go's syntax is inspired by C.",
  "Go's community is very active.",
  "Go's compiler is fast.",
  "Go's error handling promotes clarity.",
  "Go's standard library is extensive.",
  "Go's goroutines are lightweight.",
  "Go's channels facilitate communication between goroutines.",
  "Go's build process is simple.",
  "Go's package system is easy to use.",
  "Go's performance is close to C.",
  "Go's syntax is easy to learn.",
  "Go's concurrency model is powerful.",
  "Go's standard library is well-documented.",
  "Go's error handling is straightforward.",
  "Go's tooling includes a race detector.",
  "Go's community is supportive.",
  "Go's package management is simple.",
  "Go's performance is reliable.",
  "Go's interfaces are flexible.",
  "Go's tooling is comprehensive.",
  "Go's standard library includes many utilities.",
  "Go's concurrency model is efficient.",
];
