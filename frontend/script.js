let body = document.querySelector('body');


function getUrlAndUpdateElement() {
    const val = document.getElementById("url-input").value;
    // const apiUrl = "https://crumburl.fly.dev"
    const apiUrl = "http://127.0.0.1:8090"
    fetch(`${apiUrl}/api/createurl`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
        },
        body: `url=${val}`
    })
    .then(response => {
        if (!response.ok) {
            throw new Error("Invalid request")
        }
        return response
    })
    .then(response => response.text())
    .then(data => {
        element = document.getElementById("url-display");
        const component = `<a id="url-display" href="${apiUrl}/${data}">${data}</a>`
        element.outerHTML = component
        document.getElementById("copy-button").classList.remove("invisible")
    })
    .catch(err => {
        const component = `<div id="url-display" class="alert alert-danger mt-2" role="alert">
                Invalid request
        </div>`

        element = document.getElementById("url-display");
        document.getElementById("copy-button").classList.add("invisible")
        element.outerHTML = component
        console.log(err)
    })
}

function urlToClipboard() {
    const val = document.getElementById("url-display").href
    console.log("val", val)
    navigator.clipboard.writeText(val)
}

document.getElementById("submit-url").addEventListener("click", getUrlAndUpdateElement)
document.getElementById("copy-button").addEventListener("click", urlToClipboard)
