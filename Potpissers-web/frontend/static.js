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
const eventSource = new EventSource("/api/sse/main")
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
            // TODO -> each server
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = "(" + data.server_name + ") " + data.death_message
                return p
            })
            break
        }
        case "events": {
            // TODO -> each server
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = data.cap_message
                return p
            })
            break
        }
        case "chat": {
            // TODO each server
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = data.message
                return p
            })
            break
        }
        case "online": {
            console.log("hi")
            const data = jsonData.data

            const option = document.createElement("option")
            option.textContent = data.name
            option.disabled = true
            {
                const gameModeName = data.game_mode_name
                const select = document.getElementById("onlineplayers-" + gameModeName)
                select.appendChild(option)
                document.getElementById("online-" + gameModeName).innerText = "/" + gameModeName + ": " + (select.children.length - 1).toString()
            }
            {
                const select = document.getElementById("onlineplayers")
                select.appendChild(option.cloneNode(true))
                document.getElementById("online").innerText = "potpissers: " + (select.children.length - 1).toString()
            }
            break
        }
        case "offline": {
            console.log("hey")
            const data = jsonData.data
            {
                const gameModeName = data.game_mode_name
                const select = document.getElementById("onlineplayers-" + gameModeName)
                for (const option of select.querySelectorAll("option")) {
                    if (option.textContent.trim() === data.name) {
                        option.remove()

                        document.getElementById("online-" + gameModeName).textContent = "/" + gameModeName + ": " + (select.children.length - 1)
                        break
                    }
                }
            }
            {
                const select = document.getElementById("onlineplayers")
                for (const option of select.querySelectorAll("option")) {
                    if (option.textContent.trim() === data.name) {
                        option.remove()

                        document.getElementById("online").textContent = "potpissers: " + (select.children - 1)
                        break
                    }
                }
            }
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
        case "general": {
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = data.content
                return p
            })
            break
        }
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