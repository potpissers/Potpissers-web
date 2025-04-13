"use strict"
function handleClassHiddenToggle(className) {
    const elements = document.getElementsByClassName(className)
    const newValue = !elements[0].hidden
    for (let i = 0; i < elements.length; i++)
        elements[i].hidden = newValue
}
function handleTipsButtonSelect(clickedOption) {
    for (let buttonId of tipIds) {
        document.getElementById(buttonId).hidden = true
    }
    document.getElementById(clickedOption.value).hidden = false
}

function handleLineItemButtonContentMaximize() {
    if (document.getElementById("content").style.gridTemplateRows !== "auto 22vh") {
        document.getElementById("content").style.gridTemplateRows = "auto 22vh"
        document.body.style.gridTemplateRows = "44vh auto auto"
    }
}
function handleContentMaximizeButtonClick(isAnnouncements) {
    const contentElement = document.getElementById("content")
    if (isAnnouncements) {
        if (contentElement.style.gridTemplateRows === "auto auto") {
            contentElement.style.gridTemplateRows = "1fr 1fr"
            document.body.style.gridTemplateRows = "44vh 44vh auto"

            for (const element of document.getElementsByClassName("line-items"))
                element.hidden = true

            for (const element of document.querySelectorAll(".contentbody, .contenttitle"))
                element.classList.remove("h")
        } else {
            contentElement.style.gridTemplateRows = "auto auto"
            document.body.style.gridTemplateRows = "44vh auto auto"

            for (const element of document.getElementsByClassName("contentbody"))
                element.classList.add("h")
        }
    } else {
        if (contentElement.style.gridTemplateRows === "auto auto") {
            contentElement.style.gridTemplateRows = "1fr 1fr"
            document.body.style.gridTemplateRows = "44vh 44vh auto"

            for (const element of document.querySelectorAll(".contentbody, .contenttitle"))
                element.classList.remove("h")
        } else {
            contentElement.style.gridTemplateRows = "auto auto"
            document.body.style.gridTemplateRows = "44vh auto auto"

            for (const element of document.getElementsByClassName("contenttitle"))
                element.classList.add("h")
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
function handlePlayersListToggle(button) {
    switch (button.textContent) {
        case "/list":
            button.textContent = "/glist"
            document.getElementById("chat").hidden = true
            document.getElementById("onlineplayers-server").hidden = true
            document.getElementById("onlineplayers-network").hidden = false
            break
        case "/glist":
            button.textContent = "chat"
            document.getElementById("chat").hidden = false
            document.getElementById("onlineplayers-server").hidden = true
            document.getElementById("onlineplayers-network").hidden = true
            break
        case "chat":
            button.textContent = "/list"
            document.getElementById("chat").hidden = true
            document.getElementById("onlineplayers-server").hidden = false
            document.getElementById("onlineplayers-network").hidden = true
            break
    }
}

function handleSpanButtonClick(span, runnable, className) {
    span.querySelector("input").checked = true
    runnable(span, className)
}
function handleEventsButtonClick(span, className) {
    for (const event of document.getElementsByClassName(className))
        event.hidden = true
    document.getElementById(span.querySelector('button').textContent).hidden = false
}