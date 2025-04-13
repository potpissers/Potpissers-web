"use strict"
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
    let contentElement = document.getElementById("content")
    if (isAnnouncements) {
        if (contentElement.style.gridTemplateRows === "auto 22vh") {
            contentElement.style.gridTemplateRows = "1fr 1fr"
            document.body.style.gridTemplateRows = "44vh 44vh auto"

            for (const className of clickedLineItemClasses)
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

            for (const element of document.getElementsByClassName("contenttitle"))
                element.classList.remove("h")
            document.getElementById("content").style.gridTemplateRows = "1fr 1fr"
        } else {
            contentElement.style.gridTemplateRows = "22vh auto"
            document.body.style.gridTemplateRows = "44vh auto auto"

            for (const element of document.getElementsByClassName("contenttitle"))
                element.classList.add("h")
            document.getElementById("content").style.gridTemplateRows = "auto auto"
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