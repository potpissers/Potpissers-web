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
    fetch("https://potpissers.com/api/proxy/mojang/username/" + inputElement.value, {signal: nameCheckController.signal})
        .then(res => {
            inputElement.classList.add(res.status !== 404 ? "input-valid" : "input-invalid")
        })
        .catch(err => {
        }); // TODO -> warning when hasn't played before
}

function handleMcNameKeyDown(event) {
    if (event.key === "Enter")
        event.target.blur()
    else
        event.target.classList.remove("input-valid", "input-invalid");
}

function handleContentMaximizeButtonClick(isAnnouncements) {
    let contentElement = document.getElementById("content")
    if (isAnnouncements) {
        if (contentElement.style.gridTemplateRows === "auto 22vh") {
            contentElement.style.gridTemplateRows = "1fr 1fr"
            document.body.style.gridTemplateRows = "44vh 44vh auto"
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

function handleRedditVideos() { // TODO -> server handle this
    fetch("https://www.reddit.com/r/potpissers/new.json?limit=100")
        .then(response => response.json()
            .then(data => {
                const ul = document.getElementById("videos")
                data.data.children
                    .forEach(post => {
                        const url = post.data.url
                        if (url && (url.includes("youtube.com") || url.includes("youtu.be"))) {
                            const img = document.createElement("img")
                            img.src = "https://img.youtube.com/vi/" + url.match(/[?&]v=([a-zA-Z0-9_-]{11})/)[1] + "/hqdefault.jpg"
                            img.style.width = "50%"


                            const a = document.createElement("a")
                            a.href = "https://www.reddit.com" + post.data.permalink
                            a.textContent = post.data.title

                            const li = document.createElement("li")
                            li.classList.add("fsb")
                            li.appendChild(img)
                            li.appendChild(a)
                            ul.appendChild(li)
                        }
                    })
            }));
}