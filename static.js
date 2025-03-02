function handleTipsButtonClick(otherButtonsIds, clickedId) {
    for (let buttonId of otherButtonsIds) {
        document.getElementById(buttonId).hidden = true
    }

    document.getElementById(clickedId).hidden = false
}

function handleContentMaximizeButtonClick(isAnnouncements) {
    let contentElement = document.getElementById("content")
    if (isAnnouncements) {
        if (contentElement.style.gridTemplateRows === "auto 1fr") {
            contentElement.style.gridTemplateRows = "1fr 1fr"
            document.body.style.gridTemplateRows = "44vh 44vh auto"
        }
        else {
            contentElement.style.gridTemplateRows = "auto 1fr"
            document.body.style.gridTemplateRows = "44vh auto auto"
        }
    }
    else {
        if (contentElement.style.gridTemplateRows === "1fr auto") {
            contentElement.style.gridTemplateRows = "1fr 1fr"
            document.body.style.gridTemplateRows = "44vh 44vh auto"
        }
        else {
            contentElement.style.gridTemplateRows = "1fr auto"
            document.body.style.gridTemplateRows = "44vh auto auto"
        }
    }
}
function handleRedditVideos() {
    fetch("https://www.reddit.com/r/potpissers/new.json?limit=100")
        .then(response => response.json()
            .then(data => {
                const ul = document.getElementById("videos")
                data.data.children
                    .forEach(post => {
                        const url = post.data.url
                        if (url.includes("youtube.com") || url.includes("youtu.be")) {
                            const a = document.createElement("a")
                            a.textContent = post.data.title
                            a.href = url

                            const li = document.createElement("li")
                            li.appendChild(a)
                            ul.appendChild(li)
                        }
                    })
            }));
}