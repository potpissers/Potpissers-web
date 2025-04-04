"use strict"
const privateJsonLineItems = []
const currentLineItemsLowercaseUsernames = new Set()
const currentLineItemsUsernameCheckPromises = []
let currentLineItemsCost = 0

function handleAddLineItemJson(itemName, itemAmountString) {
    const username = document.getElementById("donate-username").value
    if (username.length === 0)
        return alert("invalid username: " + username)
    else {
        currentLineItemsCost += currentPrices[itemName] * parseInt(itemAmountString)
        document.getElementById("checkoutbalance").innerText = "$" + currentLineItemsCost / 100

        privateJsonLineItems.push({
            username: username,
            line_item_name: itemName,
            line_item_amount: parseInt(itemAmountString, 10),
        })
        document.getElementById("checkout").classList.remove("h")
        document.getElementById("donatebutton").hidden = true

        document.getElementById("donatesidebutton").hidden = true
        document.getElementById("donatesidebuttonred").hidden = false

        if (!currentLineItemsLowercaseUsernames.has(username.toLowerCase())) {
            currentLineItemsLowercaseUsernames.add(username.toLowerCase())
            handleUsernameApiCheck(username)
        }
    }
}
function handleUsernameApiCheck(username) {
    currentLineItemsUsernameCheckPromises.push(fetch("/api/proxy/mojang/username/" + username)
        .then(res => {
            switch (res.status) {
                case 200:
                    break
                case 404:
                    handleLineItemReset()
                    return alert("invalid username: " + username)
                default:
                    handleUsernameApiCheck(username)
            }
        })
        .catch(_ => {
            handleUsernameApiCheck(username)
        }));
}
function handlePaymentLink() {
    document.getElementById("squarebutton").classList.remove("h")

    document.getElementById("checkout").classList.add("h")
    document.getElementById("donatesidebutton").hidden = false
    document.getElementById("donatesidebuttonred").hidden = true

    Promise.all(currentLineItemsUsernameCheckPromises)
        .then(_ => {
            const json = JSON.stringify(privateJsonLineItems)
            doLineItemDataReset()

            fetch("/api/donations/payments", {
                method: "POST", body: json,
            })
                .then(response => response.text())
                .then(url => {
                    const a = document.getElementById("squarelink")
                    a.innerText = url
                    a.href = url
                    a.onclick = () => handleLineItemReset()
                })
        })
}
function doLineItemDataReset() {
    privateJsonLineItems.length = 0
    currentLineItemsLowercaseUsernames.clear()
    currentLineItemsUsernameCheckPromises.length = 0
    currentLineItemsCost = 0
}
function handleLineItemReset() {
    doLineItemDataReset()

    document.getElementById("checkoutbalance").innerText = ""
    document.getElementById("checkout").classList.add("h")
    document.getElementById("squarelink").classList.add("h")
    document.getElementById("donatebutton").hidden = false

    document.getElementById("donatesidebutton").hidden = false
    document.getElementById("donatesidebuttonred").hidden = true
}