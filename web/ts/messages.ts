const showErrorMessage = (msg) => {
    let messagesDiv = document.getElementById("messages");
    let errorMsg = document.getElementById("error-message");
    let successMsg = document.getElementById("success-message");

    messagesDiv.style.display = "inherit";

    successMsg.innerText = "";
    errorMsg.innerText = msg;
}

const showMessage = (msg) => {
    let messagesDiv = document.getElementById("messages");
    let errorMsg = document.getElementById("error-message");
    let successMsg = document.getElementById("success-message");

    messagesDiv.style.display = "inherit";

    successMsg.innerText = msg;
    errorMsg.innerText = "";
}