import {Endpoints} from "./endpoints.js";
import {YeetFileDB} from "./db.js";
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
    }).then(async response => {
        if (response.ok) {
            await resetKeys();
            showMessage("Your email has been verified! Redirecting...", false);
            setTimeout(() => {
                window.location.assign(Endpoints.HTMLSend.path)
            }, 1500);
        } else {
            fieldset.disabled = false;

        }
    })
}

const resetKeys = async () => {
    let db = new YeetFileDB();
    db.getVaultKeyPair("", true).then(async ([privKey, pubKey]) => {
        await db.removeKeys(async success => {
            if (!success) {
                alert("Error resetting vault keys!");
                return;
            }

            // Re-insert with new auth-based db file
            const dbModule = await import("./db.js?_=" + Date.now());
            let db = new dbModule.YeetFileDB();
            privKey = privKey as Uint8Array;
            pubKey = pubKey as Uint8Array;

            await db.insertVaultKeyPair(privKey, pubKey, "", success => {
                if (!success) {
                    alert("Error setting vault keys!");
                    return;
                }
            });
        });
    }).catch(e => {
        console.error(e);
        alert(e);
    });
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}