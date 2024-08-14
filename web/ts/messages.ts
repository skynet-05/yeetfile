const showMessage = (msg: string, isError: boolean) => {
    let messagesDiv = document.getElementById("messages");
    let errorMsg = document.getElementById("error-message");
    let successMsg = document.getElementById("success-message");

    if (msg.length === 0) {
        messagesDiv.style.display = "none";
        return;
    }

    messagesDiv.style.display = "inherit";

    if (isError) {
        successMsg.innerText = "";
        errorMsg.innerText = msg;
    } else {
        successMsg.innerText = msg;
        errorMsg.innerText = "";
    }
}

const clearMessages = () => {
    document.getElementById("messages").style.display = "none";
}