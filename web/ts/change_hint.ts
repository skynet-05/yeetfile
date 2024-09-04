import {MaxHintLen} from "./constants.js";
import {Endpoints} from "./endpoints.js";
import {ChangePasswordHint} from "./interfaces.js";

let hintInput: HTMLInputElement, submitHint: HTMLInputElement;

const init = () => {
    hintInput = document.getElementById("password-hint") as HTMLInputElement;
    submitHint = document.getElementById("submit-pw-hint") as HTMLInputElement;

    submitHint.addEventListener("click", submitPasswordHint);
    document.addEventListener("keydown", (event: KeyboardEvent) => {
        if (event.key === "Enter") {
            submitHint.click();
        }
    });
}

const inputsDisabled = (disabled: boolean) => {
    hintInput.disabled = disabled;
    submitHint.disabled = disabled;
}

const submitPasswordHint = () => {
    if (hintInput.value.length > MaxHintLen) {
        alert(`Hint exceeds max length (${MaxHintLen})`);
        return;
    }

    inputsDisabled(true);
    let changeHint = new ChangePasswordHint();
    changeHint.hint = hintInput.value;

    fetch(Endpoints.ChangeHint.path, {
        method: "POST",
        body: JSON.stringify(changeHint, jsonReplacer)
    }).then(async response => {
        if (response.ok) {
            alert("Password hint updated!");
            window.location.assign(Endpoints.HTMLAccount.path);
        } else {
            inputsDisabled(false);
            showMessage("Error changing hint: " + await response.text(), true);
        }
    }).catch(error => {
        inputsDisabled(false);
        console.error(error);
        alert("Error updating password hint!");
    });
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}