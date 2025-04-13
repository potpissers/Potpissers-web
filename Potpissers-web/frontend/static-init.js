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
    if (document.getElementById("content").style.gridTemplateRows !== "auto auto") {
        document.getElementById("content").style.gridTemplateRows = "auto auto"
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
    switch (button.textContent.trim()) {
        case "game":
            button.textContent = "discord"
            for (const element of document.getElementsByClassName("id-chat-discord"))
                element.hidden = false
            for (const element of document.getElementsByClassName("id-chat-game"))
                element.hidden = true
            break
        case "discord":
            button.textContent = "game"
            for (const element of document.getElementsByClassName("id-chat-discord"))
                element.hidden = true
            for (const element of document.getElementsByClassName("id-chat-game"))
                element.hidden = false
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