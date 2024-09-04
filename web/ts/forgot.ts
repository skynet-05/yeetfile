import {ForgotPassword} from "./interfaces.js";
import {Endpoints} from "./endpoints.js";

let emailInput: HTMLInputElement, submitBtn: HTMLInputElement;

const inputsDisabled = (disabled: boolean) => {
    emailInput.disabled = disabled;
    submitBtn.disabled = disabled;
}

const init = () => {
    emailInput = document.getElementById("email-address") as HTMLInputElement;
    submitBtn = document.getElementById("submit") as HTMLInputElement;
    submitBtn.addEventListener("click", () => {
        inputsDisabled(true);

        let email = emailInput.value;
        if (email.indexOf("@") <= 0 || email.indexOf(".") <= 0) {
            showMessage("Invalid email", true);
            return;
        }

        let forgotPassword = new ForgotPassword();
        forgotPassword.email = email;

        fetch(Endpoints.Forgot.path, {
            method: "POST",
            body: JSON.stringify(forgotPassword, jsonReplacer)
        }).then(async response => {
            if (response.ok) {
                showMessage("Your request has been submitted. If a " +
                    "password hint was set for this account, you will receive " +
                    "an email shortly. If you don't receive an email after a " +
                    "few minutes and it isn't in your spam folder, contact " +
                    "the host for assistance.", false);
            } else {
                inputsDisabled(false);
                showMessage("Error submitting request: " + await response.text(), true);
            }
        }).catch(error => {
            inputsDisabled(false);
            console.error(error);
            alert("Error submitting request!");
        });
    });
}

if (document.readyState !== 'loading') {
    init();
} else {
    document.addEventListener('DOMContentLoaded', () => {
        init();
    });
}