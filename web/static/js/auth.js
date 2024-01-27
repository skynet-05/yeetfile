document.addEventListener("DOMContentLoaded", () => {
    setupToggles();
});

const setupToggles = () => {
    let emailToggle = document.getElementById("email-signup");
    let emailDiv = document.getElementById("email-div");

    let idToggle = document.getElementById("id-signup");
    let idDiv = document.getElementById("account-id-div");

    emailToggle.addEventListener("click", () => {
        emailDiv.style.display = "inherit";
        idDiv.style.display = "none";
    });

    idToggle.addEventListener("click", () => {
        emailDiv.style.display = "none";
        idDiv.style.display = "inherit";
    });
}

const addAuthError = (msg) => {
    let messagesDiv = document.getElementById("messages");
    let errorMsg = document.getElementById("error-message");
    let successMsg = document.getElementById("success-message");

    messagesDiv.style.display = "inherit";

    successMsg.innerText = "";
    errorMsg.innerText = msg;
}

const addAuthMessage = (msg) => {
    let messagesDiv = document.getElementById("messages");
    let errorMsg = document.getElementById("error-message");
    let successMsg = document.getElementById("success-message");

    messagesDiv.style.display = "inherit";

    successMsg.innerText = msg;
    errorMsg.innerText = "";
}

const addAuthHTML = (html) => {
    let messagesDiv = document.getElementById("messages");
    let errorMsg = document.getElementById("error-message");
    let successMsg = document.getElementById("success-message");

    messagesDiv.style.display = "inherit";

    successMsg.innerText = "";
    errorMsg.innerText = "";

    messagesDiv.innerHTML = html;
}