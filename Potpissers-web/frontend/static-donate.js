"use strict"
const privateJsonLineItems = []
const currentLineItemsLowercaseUsernames = new Set()
const currentLineItemsUsernameCheckPromises = []
let currentLineItemsCost = 0

function handleAddLineItemJson(usernameInputString, itemName, itemAmountString) {
    if (usernameInputString.length === 0)
        return alert("invalid username: " + usernameInputString)
    else {
        currentLineItemsCost += currentPrices[itemName] * parseInt(itemAmountString)
        document.getElementById("checkoutbalance").innerText = "$" + currentLineItemsCost / 100

        privateJsonLineItems.push({
            username: usernameInputString,
            line_item_name: itemName,
            line_item_amount: parseInt(itemAmountString, 10),
        })
        document.getElementById("checkout").classList.remove("h")
        document.getElementById("donatebutton").hidden = true

        document.getElementById("donatesidebutton").hidden = true
        document.getElementById("donatesidebuttonred").hidden = false

        if (!currentLineItemsLowercaseUsernames.has(usernameInputString.toLowerCase())) {
            currentLineItemsLowercaseUsernames.add(usernameInputString.toLowerCase())
            handleUsernameApiCheck(usernameInputString)
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

    const squareLink = document.getElementById("squarelink")
    squareLink.innerText = ""
    squareLink.href = ""

    Promise.all(currentLineItemsUsernameCheckPromises)
        .then(_ => {
            const json = JSON.stringify(privateJsonLineItems)
            doLineItemDataReset()

            fetch("/api/donations/payments", {
                method: "POST", body: json,
            })
                .then(response => response.text())
                .then(url => {
                    squareLink.hidden = false
                    document.getElementById("squarelinkspinner").hidden = true
                    squareLink.innerText = url
                    squareLink.href = url
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
    document.getElementById("squarebutton").classList.add("h")
    document.getElementById("squarelink").hidden = true
    document.getElementById("squarelinkspinner").hidden = false
    document.getElementById("donatebutton").hidden = false

    document.getElementById("donatesidebutton").hidden = false
    document.getElementById("donatesidebuttonred").hidden = true
}