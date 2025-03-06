function handleTipsButtonClick(otherButtonsIds, clickedId) {
    for (let buttonId of otherButtonsIds) {
        document.getElementById(buttonId).hidden = true
    }

    document.getElementById(clickedId).hidden = false
}
function handleClassHiddenToggle(className) {
    document.getElementById("donateusername").hidden = false

    const elements = document.getElementsByClassName(className)
    const newValue = !elements[0].hidden
    for (let i = 0; i < elements.length; i++)
        elements[i].hidden = newValue
}
let nameCheckController = null;
function handleMcNameCheck(inputElement) {
    if (nameCheckController)
        nameCheckController.abort()

    const username = inputElement.value
    if (username === "")
        inputElement.classList.remove("input-valid", "input-invalid");
    else {
        nameCheckController = new AbortController()
        fetch("https://potpissers.com/api/proxy/mojang/username/" + username, { signal: nameCheckController.signal })
            .then(res => {
                inputElement.classList.add(res.status !== 404 ? "input-valid" : "input-invalid")
            })
            .catch(err => {
                if (err.name !== "AbortError")
                    console.error(err.message)
            });
    }
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
const privateJsonLineItems = []
function handleAddLineItemJson(itemName, itemAmountString) {
    privateJsonLineItems.push({
        username: document.getElementById("donateusername").value,
        line_item_name: itemName,
        line_item_amount: parseInt(itemAmountString, 10),
    })
    document.getElementById("checkout").hidden = false
}
function fetchPaymentLink() {
    fetch("https://potpissers.com/api/donate", {
        method: "POST", body: JSON.stringify(privateJsonLineItems),
    })
        .then(response => response.text())
        .then(url => window.open(url, "_blank"))
}