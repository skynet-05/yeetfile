import {Endpoints} from "./endpoints.js";
import {NewTOTP, SetTOTP, SetTOTPResponse} from "./interfaces.js";

const init = () => {
    let loadingDiv = document.getElementById("loading-div");
    fetch(Endpoints.TwoFactor.path)
        .then(response => {
            loadingDiv.className = "hidden";
            if (!response.ok) {
                showMessage("Error fetching 2FA values", true);
                return;
            }

            response.json().then(json => {
                let totp = new NewTOTP(json);
                showTOTP(totp);
        });
    })
}

const showTOTP = (totp: NewTOTP) => {
    let twoFactorDiv = document.getElementById("2fa-div");
    twoFactorDiv.className = ""; // Removes "hidden" class

    let qrImg = document.getElementById("2fa-qr") as HTMLImageElement;
    let qrSecret = document.getElementById("2fa-secret");
    let codeInput = document.getElementById("2fa-code") as HTMLInputElement;
    let submitBtn = document.getElementById("submit-2fa") as HTMLInputElement;

    qrImg.src = `data:image/jpeg;base64,${totp.b64Image}`
    qrSecret.innerText = `TOTP Secret: ${totp.secret}`

    submitBtn.addEventListener("click", () => {
        if (codeInput.value.length !== 6) {
            alert("Code input must be 6-digits long");
            return;
        }

        clearMessages();

        submitBtn.disabled = true;
        codeInput.disabled = true;

        let setTOTP = new SetTOTP();
        setTOTP.secret = totp.secret;
        setTOTP.code = codeInput.value;

        fetch(Endpoints.TwoFactor.path, {
            method: "POST",
            body: JSON.stringify(setTOTP, jsonReplacer)
        }).then(response => {
            if (!response.ok) {
                submitBtn.disabled = false;
                codeInput.disabled = false;
                response.text().then(text => {
                    showMessage(`Error: ${text} -- 
                    Please double check your 2FA code`, true);
                })
                return;
            }

            response.json().then(json => {
                let totpResponse = new SetTOTPResponse(json);
                showRecovery(totpResponse);
            });
        })
    });
}

const showRecovery = (response: SetTOTPResponse) => {
    let twoFactorDiv = document.getElementById("2fa-div");
    twoFactorDiv.className = "hidden";

    let recoveryDiv = document.getElementById("recovery-div");
    recoveryDiv.className = ""; // Removes "hidden" class

    let recoveryCodes = document.getElementById("recovery-codes");

    for (let i in response.recoveryCodes) {
        recoveryCodes.innerText += response.recoveryCodes[i] + "\r\n";
    }
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}