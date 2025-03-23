const privateJsonLineItems = []
let currentLineItemsCost = 0
function handleAddLineItemJson(itemName, itemAmountString) {
    currentLineItemsCost += currentPrices[itemName] * parseInt(itemAmountString)
    document.getElementById("checkoutbalance").innerText = "$" + currentLineItemsCost / 100

    privateJsonLineItems.push({
        username: document.getElementById("donate-username").value,
        line_item_name: itemName,
        line_item_amount: parseInt(itemAmountString, 10),
    })
    document.getElementById("checkout").classList.remove("h")
    document.getElementById("donatebutton").hidden = true
    document.getElementById("donatesidebutton").classList.add("r")
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
    document.getElementById("donatesidebutton").classList.remove("r")
}