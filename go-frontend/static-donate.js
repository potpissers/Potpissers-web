"use strict"
const privateJsonLineItems = []
let currentLineItemsCost = 0
function handleAddLineItemJson(itemName, itemAmountString) {
    currentLineItemsCost += currentPrices[itemName] * parseInt(itemAmountString)
    document.getElementById("checkoutbalance").innerText = "$" + currentLineItemsCost / 100

    const username = document.getElementById("donate-username").value
    privateJsonLineItems.push({
        username: username,
        line_item_name: itemName,
        line_item_amount: parseInt(itemAmountString, 10),
    })
    document.getElementById("checkout").classList.remove("h")
    document.getElementById("donatebutton").hidden = true

    document.getElementById("donatesidebutton").hidden = true
    document.getElementById("donatesidebuttonred").hidden = false

    fetch("/api/proxy/mojang/username/" + username)
        .then(res => {
            if (res.status === 404)
                doLineItemReset()
        })
        .catch(err => {
        });
}

function fetchPaymentLink() {
    const json = JSON.stringify(privateJsonLineItems)
    doLineItemReset()

    fetch("/api/donations/payments", {
        method: "POST", body: json,
    })
        .then(response => response.text())
        .then(url => window.open(url, "_blank"))
}
function doLineItemReset() {
    privateJsonLineItems.length = 0
    currentLineItemsCost = 0

    document.getElementById("checkoutbalance").innerText = ""
    document.getElementById("checkout").classList.add("h")
    document.getElementById("donatebutton").hidden = false

    document.getElementById("donatesidebutton").hidden = false
    document.getElementById("donatesidebuttonred").hidden = true
}