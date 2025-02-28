function handleTipsButtonClick(otherButtonsIds, clickedId) {
    for (let buttonId of otherButtonsIds) {
        document.getElementById(buttonId).hidden = true
    }

    document.getElementById(clickedId).hidden = false
}

function handleContentMaximizeButtonClick(isAnnouncements) {
    document.getElementById("content").style.gridTemplateRows = isAnnouncements ? "auto 1fr" : "1fr auto"
}