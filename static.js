"use strict"
function handleTipsButtonClick(otherButtonsIds, clickedId) {
    for (let buttonId of otherButtonsIds) {
        document.getElementById(buttonId).hidden = true
    }

    document.getElementById(clickedId).hidden = false
}

function handleClassHiddenToggle(className) {
    const elements = document.getElementsByClassName(className)
    const newValue = !elements[0].hidden
    for (let i = 0; i < elements.length; i++)
        elements[i].hidden = newValue
}

let nameCheckController = null;
function handleMcNameBlur(inputElement) {
    if (nameCheckController)
        nameCheckController.abort()

    nameCheckController = new AbortController()
    fetch("/api/proxy/mojang/username/" + inputElement.value, {signal: nameCheckController.signal})
        .then(res => {
            inputElement.classList.add(res.status !== 404 ? "v" : "iv")
        })
        .catch(err => {
        }); // TODO -> warning when hasn't played before
}

function handleMcNameKeyDown(event) {
    if (event.key === "Enter")
        event.target.blur()
    else
        event.target.classList.remove("v", "iv");
}

function handleLineItemButtonContentMaximize() {
    if (document.getElementById("content").style.gridTemplateRows !== "auto 22vh") {
        document.getElementById("content").style.gridTemplateRows = "auto 22vh"
        document.body.style.gridTemplateRows = "44vh auto auto"
    }
}
function handleContentMaximizeButtonClick(isAnnouncements) {
    let contentElement = document.getElementById("content")
    if (isAnnouncements) {
        if (contentElement.style.gridTemplateRows === "auto 22vh") {
            contentElement.style.gridTemplateRows = "1fr 1fr"
            document.body.style.gridTemplateRows = "44vh 44vh auto"
            // TODO -> reset line item classes
        } else {
            contentElement.style.gridTemplateRows = "auto 22vh"
            document.body.style.gridTemplateRows = "44vh auto auto"
        }
    } else {
        if (contentElement.style.gridTemplateRows === "22vh auto") {
            contentElement.style.gridTemplateRows = "1fr 1fr"
            document.body.style.gridTemplateRows = "44vh 44vh auto"
        } else {
            contentElement.style.gridTemplateRows = "22vh auto"
            document.body.style.gridTemplateRows = "44vh auto auto"
        }
    }
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
            handleSseLi(jsonData, (li, data) => {
                const p = document.createElement("p")
                p.textContent = data.name
                return p
            })
            const onlineNumber = document.getElementById("online-server")
            onlineNumber.textContent = (parseInt(onlineNumber.textContent) + 1).toString()
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
function handleChatToggle(button) {
    switch (button.textContent) {
        case "game":
            button.textContent = "discord"
            Array.from(document.getElementsByClassName("id-chat-discord"))
                .forEach(each => each.hidden = false)
            Array.from(document.getElementsByClassName("id-chat-game"))
                .forEach(each => each.hidden = true)
            break
        case "discord":
            button.textContent = "game"
            Array.from(document.getElementsByClassName("id-chat-discord"))
                .forEach(each => each.hidden = true)
            Array.from(document.getElementsByClassName("id-chat-game"))
                .forEach(each => each.hidden = false)
            break
    }
}