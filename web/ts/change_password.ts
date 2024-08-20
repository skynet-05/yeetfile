import * as crypto from "./crypto.js";
import {Endpoints} from "./endpoints.js";
import * as interfaces from "./interfaces.js";

let submitBtn: HTMLButtonElement;
let inputFields: HTMLFieldSetElement;

const init = () => {
    inputFields = document.getElementById("input-fields") as HTMLFieldSetElement;
    submitBtn = document.getElementById("change-pw-btn") as HTMLButtonElement;
    submitBtn.addEventListener("click", submitPasswordChange);

    document.addEventListener("keydown", (event: KeyboardEvent) => {
        if (event.key === "Enter") {
            submitBtn.click();
        }
    });
}

const inputsDisabled = (disabled: boolean) => {
    submitBtn.disabled = disabled;
    inputFields.disabled = disabled;
}

const submitPasswordChange = async () => {
    let id = document.getElementById("identifier") as HTMLInputElement;
    let oldPw = document.getElementById("password") as HTMLInputElement;
    let newPw = document.getElementById("new-password") as HTMLInputElement;
    let newPwConfirm = document.getElementById("new-password-confirm") as HTMLInputElement;

    if (newPw.value !== newPwConfirm.value) {
        showMessage("Passwords don't match", true);
        return;
    }

    inputsDisabled(true);
    let protectedKey: Uint8Array;
    try {
        let protectedKeyResponse = await fetch(Endpoints.ProtectedKey.path);
        let responseData = await protectedKeyResponse.json();
        protectedKey = new interfaces.ProtectedKeyResponse(responseData).protectedKey;
    } catch (error) {
        console.error(error);
        inputsDisabled(false);
        showMessage("Error fetching protected key", true);
        return;
    }

    let oldLoginKeyHash, newLoginKeyHash, newProtectedKey;
    try {
        let oldUserKey = await crypto.generateUserKey(id.value, oldPw.value);
        oldLoginKeyHash = await crypto.generateLoginKeyHash(oldUserKey, oldPw.value);

        let privateKey = await crypto.decryptChunk(oldUserKey, protectedKey);

        let newUserKey = await crypto.generateUserKey(id.value, newPw.value);
        newLoginKeyHash = await crypto.generateLoginKeyHash(newUserKey, newPw.value);

        newProtectedKey = await crypto.encryptChunk(newUserKey, privateKey);
    } catch (error) {
        inputsDisabled(false);
        showMessage("Decryption error", true);
        return;
    }

    let changePassword = new interfaces.ChangePassword();
    changePassword.prevLoginKeyHash = oldLoginKeyHash;
    changePassword.newLoginKeyHash = newLoginKeyHash;
    changePassword.protectedKey = newProtectedKey;

    fetch(Endpoints.ChangePassword.path, {
        method: "PUT",
        body: JSON.stringify(changePassword, jsonReplacer)
    }).then(async response => {
        if (response.ok) {
            alert("Your password has been changed!");
            window.location.assign("/account");
        } else {
            inputsDisabled(false);
            showMessage("Error updating password: " + await response.text(), true);
        }
    }).catch(error => {
        inputsDisabled(false);
        console.error(error);
        alert("Error updating password!");
    });

}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}