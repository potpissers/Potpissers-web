"use strict"
function handleTipsButtonClick(otherButtonsIds, clickedId) {
    for (let buttonId of otherButtonsIds) {
        document.getElementById(buttonId).hidden = true
    }

    document.getElementById(clickedId).hidden = false
}

const clickedLineItemClasses = new Set()
function handleClassHiddenToggle(className) {
    const elements = document.getElementsByClassName(className)
    const newValue = !elements[0].hidden
    for (let i = 0; i < elements.length; i++)
        elements[i].hidden = newValue

    if (!newValue)
        clickedLineItemClasses.add(className)
    else
        clickedLineItemClasses.delete(className)
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

            for (let className of clickedLineItemClasses)
                handleClassHiddenToggle(className)
            clickedLineItemClasses.clear()
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