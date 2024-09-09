import * as crypto from "./crypto.js";
import {Endpoints} from "./endpoints.js";
import {ChangeEmail, ProtectedKeyResponse} from "./interfaces.js";

let identifierInput: HTMLInputElement,
    passwordInput: HTMLInputElement,
    newEmailInput: HTMLInputElement,
    submitBtn: HTMLInputElement;

let identifierDisabled: boolean;

const init = () => {
    identifierInput = document.getElementById("identifier") as HTMLInputElement;
    passwordInput = document.getElementById("password") as HTMLInputElement;
    newEmailInput = document.getElementById("new-email") as HTMLInputElement;

    identifierDisabled = identifierInput.disabled;

    submitBtn = document.getElementById("change-email-btn") as HTMLInputElement;
    submitBtn.addEventListener("click", submitChangeEmail);
    document.addEventListener("keydown", (event: KeyboardEvent) => {
        if (event.key === "Enter") {
            submitBtn.click();
        }
    });
}

const disableInputs = (disabled: boolean) => {
    identifierInput.disabled = disabled;
    passwordInput.disabled = disabled;
    newEmailInput.disabled = disabled;
    submitBtn.disabled = disabled;

    if (!disabled) {
        identifierInput.disabled = identifierDisabled;
    }
}

const submitChangeEmail = async () => {
    disableInputs(true);

    let identifier = identifierInput.value;
    let password = passwordInput.value;
    let newEmail = newEmailInput.value;

    let oldUserKey = await crypto.generateUserKey(identifier, password);
    let oldLoginKeyHash = await crypto.generateLoginKeyHash(oldUserKey, password);

    let newUserKey = await crypto.generateUserKey(newEmail, password);
    let newLoginKeyHash = await crypto.generateLoginKeyHash(newUserKey, password);

    let protectedKeyResponse = await fetch(Endpoints.ProtectedKey.path);
    let protectedKey = new ProtectedKeyResponse(
        await protectedKeyResponse.json()
    ).protectedKey;

    let newProtectedKey;
    try {
        let privateKey = await crypto.decryptChunk(oldUserKey, protectedKey);
        newProtectedKey = await crypto.encryptChunk(newUserKey, privateKey);
    } catch (e) {
        showMessage("Incorrect password", true);
        disableInputs(false);
        return;
    }

    let changeEmail = new ChangeEmail();
    changeEmail.oldLoginKeyHash = oldLoginKeyHash;
    changeEmail.newLoginKeyHash = newLoginKeyHash;
    changeEmail.protectedKey = newProtectedKey;
    changeEmail.newEmail = newEmail;

    let changeID = window.location.href.split("/").pop()
    fetch(Endpoints.format(Endpoints.ChangeEmail, changeID), {
        method: "PUT",
        body: JSON.stringify(changeEmail, jsonReplacer)
    }).then(async response => {
        if (response.ok) {
            window.location.assign(Endpoints.HTMLVerifyEmail.path + "?email=" + newEmail);
        } else {
            disableInputs(false);
            showMessage(`Error: ${await response.text()}`, true);
        }
    }).catch(() => {
        disableInputs(false);
        alert("Error submitting change email request");
        return;
    });
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}