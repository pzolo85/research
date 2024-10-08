document.body.onload = start
    const groups = [
        "basketball",
        "tennis",
        "baseball"
    ];
    
    const users = [
        "alice",
        "bob",
        "charlie",
        "david",
        "eve",
        "frank",
        "grace",
        "hannah",
        "isaac",
        "jack"
    ];
    
function start(){
    const pm = document.getElementById("msg")
    const grpm = document.getElementById("grpmsg")
    const glbm = document.getElementById("glbmsg")
    const ui = Math.floor(Math.random() * users.length)
    const gi = Math.floor(Math.random() * groups.length)
    const user = users[ui]
    const group = groups[gi]

    const t = document.getElementById("title")
    t.textContent = `Hello ${user}`
    const gt = document.getElementById("grptitle")
    gt.textContent = `Group ${group} messages:`

    const evs = new EventSource(`/sub/messages?user=${user}&group=${group}`)
    attachEvent(evs,"private",pm)
    attachEvent(evs,"group",grpm)
    attachEvent(evs,"global",glbm)

    evs.onmessage = (event) => {
        console.log("received unknown event type")
        console.log(event)
    }

    evs.onerror = (event) => {
        console.log("received an error")
        console.log(event)
    }
}

function attachEvent(es, t , node){
    es.addEventListener(t,e => {
        const m = document.createElement("li")
        m.textContent = `${e.timeStamp} | ${e.data}` 
        node.appendChild(m)
    })
}