import * as crypto from "./crypto.js";
import * as endpoints from "./endpoints.js";
import {YeetFileDB} from "./db.js";

const useVaultPasswordKey = "UseVaultPassword";
const useVaultPasswordValue = "true";

let vaultPasswordDialog;
let vaultPasswordCB;
let loginBtn;
let buttonLabel;

const init = () => {
    vaultPasswordCB = document.getElementById("vault-pass-cb");
    vaultPasswordDialog = document.getElementById("vault-pass-dialog");
    loginBtn = document.getElementById("login-btn");
    buttonLabel = loginBtn.value;
    loginBtn.addEventListener("click", async () => {
        await login();
    });

    if (localStorage.getItem(useVaultPasswordKey) === useVaultPasswordValue) {
        vaultPasswordCB.checked = true;
    } else {
        vaultPasswordCB.checked = false;
    }
}

const resetLoginButton = () => {
    let btn = document.getElementById("login-btn");
    btn.disabled = false;
    btn.value = buttonLabel;
}

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
    xhr.open("POST", endpoints.Login, false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = async () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            let response = JSON.parse(xhr.responseText);
            let privKeyBytes = base64ToArray(response.protectedKey);
            let pubKeyBytes = base64ToArray(response.publicKey);
            let decPrivKeyBytes = new Uint8Array(await crypto.decryptChunk(userKey, privKeyBytes));

            if (vaultPasswordCB.checked) {
                showVaultPassDialog(decPrivKeyBytes, pubKeyBytes);
            } else {
                localStorage.setItem(useVaultPasswordKey, "");
                await new YeetFileDB().insertVaultKeyPair(decPrivKeyBytes, pubKeyBytes, "", success => {
                    if (success) {
                        window.location = "/account";
                    } else {
                        alert("Failed to insert vault keys into indexeddb");
                        window.location = endpoints.Logout;
                    }
                });
            }
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
            showErrorMessage("Missing email or 16-digit account ID");
            loginBtn.disabled = false;
            loginBtn.value = buttonLabel;
            return false;
        }

        return true;
    }
}

const showVaultPassDialog = (privKeyBytes, pubKeyBytes) => {
    let cancel = document.getElementById("cancel-pass")
    cancel.addEventListener("click", () => {
        resetLoginButton();
        new YeetFileDB().removeKeys(success => {
            if (success) {
                fetch(endpoints.Logout).catch(() => {
                    console.warn("error logging user out");
                });
            } else {
                console.warn("error removing keys");
            }
        });
        vaultPasswordDialog.close();
    });

    let submit = document.getElementById("submit-pass");
    submit.addEventListener("click", async () => {
        localStorage.setItem(useVaultPasswordKey, useVaultPasswordValue);
        let password = document.getElementById("vault-pass").value;
        await new YeetFileDB().insertVaultKeyPair(privKeyBytes, pubKeyBytes, password);
        window.location = "/account";
    });

    vaultPasswordDialog.showModal();
}

if (document.readyState !== 'loading') {
    init();
} else {
    document.addEventListener('DOMContentLoaded', () => {
        init();
    });
}