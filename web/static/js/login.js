import * as crypto from "./crypto.js"

document.addEventListener("DOMContentLoaded", () => {
    let loginBtn = document.getElementById("login-btn");
    loginBtn.addEventListener("click", async () => {
        await login();
    })
})

const login = async () => {
    let btn = document.getElementById("login-btn");
    btn.disabled = true;
    btn.value = "Logging in...";

    let identifier = document.getElementById("identifier");
    let password = document.getElementById("password");

    if (!isValidIdentifier(identifier.value)) {
        return;
    }

    let userKey = await crypto.generateUserKey(identifier.value, password.value);
    let loginKeyHash = await crypto.generateLoginKeyHash(userKey, password.value);

    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/login", false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            let response = JSON.parse(xhr.responseText);
            crypto.ingestProtectedKey(userKey, response.protectedKey, privateKey => {
                crypto.ingestPublicKey(response.publicKey, publicKey => {
                    new YeetFileDB().insertVaultKeyPair(privateKey, publicKey);
                    window.location = "/";
                });
            });
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            showErrorMessage("Error " + xhr.status + ": " + xhr.responseText);
            btn.disabled = false;
            btn.value = "Log In";
        }
    };

    xhr.send(JSON.stringify({
        identifier: identifier.value,
        loginKeyHash: Array.from(loginKeyHash)
    }));
}

const isValidIdentifier = (identifier) => {
    if (identifier.includes("@")) {
        return true;
    } else {
        if (identifier.length !== 16) {
            showErrorMessage("Invalid email or 16-digit account ID");
            return false;
        }

        return true;
    }
}