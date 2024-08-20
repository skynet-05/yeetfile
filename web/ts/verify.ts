import {Endpoints} from "./endpoints.js";
import * as constants from "./constants.js";
import * as interfaces from "./interfaces.js";

let emailInput, codeInput, codeSubmit;

const init = () => {
    emailInput = document.getElementById("email") as HTMLInputElement;
    codeInput = document.getElementById("verify") as HTMLInputElement;
    codeSubmit = document.getElementById("submit-verify") as HTMLButtonElement;

    codeSubmit.addEventListener("click", submitVerificationCode);

    // Auto-submit if code is already filled out
    if (codeInput.value.length === constants.VerificationCodeLength) {
        codeSubmit.click();
    }

    registerEnterKeySubmit(codeSubmit);
}

const submitVerificationCode = () => {
    let fieldset = document.getElementById("verify-fieldset") as HTMLFieldSetElement;
    fieldset.disabled = true;

    let emailVerify = new interfaces.VerifyEmail();
    emailVerify.code = codeInput.value;
    emailVerify.email = emailInput.value;

    fetch(Endpoints.VerifyEmail.path, {
        method: "POST",
        body: JSON.stringify(emailVerify, jsonReplacer)
    }).then(response => {
        if (response.ok) {
            showMessage("Your email has been verified! Redirecting...", false);
            setTimeout(() => {
                window.location.assign(Endpoints.HTMLAccount.path)
            }, 1500);
        } else {
            fieldset.disabled = false;

        }
    })
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}