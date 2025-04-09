"use strict"

function handleMcNameBlur(inputElement) {
    if (inputElement.value.length === 0) {}
    // TODO -> autocomplete online players
}
function handleMcNameKeyDown(event) {
    if (event.key === "Enter")
        event.target.blur()
}

function handleSseLi(jsonData, getLiChild) {
    const data = jsonData.data

    const li = document.createElement("li")
    li.appendChild(getLiChild(li, data))

    document.getElementById(jsonData.type).appendChild(li)
}
eventSource.onerror = () => location.reload()
eventSource.onmessage = function(e) {
    const jsonData = JSON.parse(e.data)
    switch (jsonData.type) {
        case "referrals": {
            handleSseLi(jsonData, (li, data) => {
                li.setAttribute('title', data.timestamp)

                const small = document.createElement("small")
                small.textContent = data.player_name + " (" + data.referrer + ") [" + data.row_number + "]"
                return small
            })
            break
        }
        case "deaths": {
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = "(" + data.server_name + ") " + data.death_message
                return p
            })
            break
        }
        case "events": {
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = data.cap_message
                return p
            })
            break
        }
        case "chat": {
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = data.message
                return p
            })
            break
        }
        case "online": {
            const data = jsonData.data

            const li = document.createElement("li")
            const p = document.createElement("p")
            p.textContent = data.name
            li.appendChild(p)

            const gameModeName = data.game_mode_name
            const ul = document.getElementById("onlineplayers-" + gameModeName)
            ul.appendChild(li)

            console.log(gameModeName)
            document.getElementById("online-" + gameModeName).innerText = ul.children.length.toString()
            break
        }
        case "offline": {
            let flag = false
            document.getElementById("online")
                .querySelectorAll("li")
                .forEach(li => {
                    if (li.textContent === jsonData.data.name) {
                        li.remove()

                        const onlineNumber = document.getElementById("online-server")
                        onlineNumber.textContent = (parseInt(onlineNumber.textContent) - 1).toString()
                        flag = true // TODO -> don't use forEach
                    }
                })
            if (!flag)
                throw new Error("offline player not found")
        }
        case "donations": {
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = data.total_money.amount + "+" + data.total_tip_money.amount
                return p
            })
            break
        }
        case "videos": {
            handleSseLi(jsonData, (li, data) => {
                if (data.youtube_embed_url !== null) {
                    const iframe = document.createElement("iframe")
                    iframe.src = data.youtube_embed_url
                    iframe.style.width = "50%"
                    li.appendChild(iframe)
                }
                else {
                    const video = document.createElement("video")
                    video.src = data.video_url
                    video.style.width = "50%"
                    li.appendChild(video)
                }
                const a = document.createElement("a")
                a.href = data.post_url
                const small = document.createElement("small")
                small.textContent = data.title
                a.appendChild(small)
                return a
            })
        }
        case "texts": {
            // TODO impl
        }
        case "general":
        case "changelog":
        case "announcements": {
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = data.content
                return p
            })
            break
        }
    }
}