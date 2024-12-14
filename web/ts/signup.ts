import {MaxHintLen} from "./constants.js";
import * as crypto from "./crypto.js";
import {Endpoints} from "./endpoints.js";
import * as interfaces from "./interfaces.js";

let emailToggle;
let idToggle;
let serverPassword;

const verifyButtonID = "verify-account";

const init = () => {
    setupToggles();

    serverPassword = document.getElementById("server-password") as HTMLInputElement;

    // Email signup
    let emailSignupButton = document.getElementById("create-email-account") as HTMLButtonElement;
    emailSignupButton.addEventListener("click", async (event) => {
        await emailSignup();
    });

    // Account ID only signup
    let accountSignupButton = document.getElementById("create-id-only-account") as HTMLButtonElement;
    accountSignupButton.addEventListener("click", (event) => {
        accountIDOnlySignup();
    });

    // Enter key submits login form
    document.addEventListener("keydown", (event: KeyboardEvent) => {
        if (event.key !== "Enter") {
            return;
        }

        if (idToggle.checked) {
            let fieldset = document.getElementById("signup-fieldset") as HTMLFieldSetElement;
            if (fieldset.disabled) {
                document.getElementById(verifyButtonID).click();
            } else {
                accountSignupButton.click();
            }
        } else if (emailToggle.checked) {
            emailSignupButton.click();
        }
    });
};

const setupToggles = () => {
    emailToggle = document.getElementById("email-signup");
    let emailDiv = document.getElementById("email-div");

    idToggle = document.getElementById("id-signup");
    let idDiv = document.getElementById("account-id-div");

    emailToggle.addEventListener("click", () => {
        emailDiv.style.display = "inherit";
        idDiv.style.display = "none";
    });

    idToggle.addEventListener("click", () => {
        emailDiv.style.display = "none";
        idDiv.style.display = "inherit";
    });
}

/**
 * generateKeys generates the necessary keys for using YeetFile
 * @param identifier {string} - either email or account ID
 * @param password {string} - the user's password
 * @param keyCallback {(Uint8Array, Uint8Array)} - the user's new private and public keys
 * @returns {Promise<interfaces.Signup>}
 */
const generateKeys = async (
    identifier: string,
    password: string,
    keyCallback: (
        signup: interfaces.Signup,
        privKey: Uint8Array,
        pubKey: Uint8Array,
    ) => void,
) => {
    let userKey = await crypto.generateUserKey(identifier, password);
    let loginKeyHash = await crypto.generateLoginKeyHash(userKey, password);
    let keyPair = await crypto.generateKeyPair();
    let publicKey = await crypto.exportKey(keyPair.publicKey, "spki");
    let privateKey = await crypto.exportKey(keyPair.privateKey, "pkcs8");
    let protectedPrivateKey = await crypto.encryptChunk(userKey, privateKey);

    let vaultFolderKey = await crypto.generateRandomKey();
    let protectedVaultFolderKey = await crypto.encryptRSA(keyPair.publicKey, vaultFolderKey);

    let signup = new interfaces.Signup();
    signup.loginKeyHash = loginKeyHash;
    signup.publicKey = publicKey;
    signup.protectedPrivateKey = protectedPrivateKey;
    signup.protectedVaultFolderKey = protectedVaultFolderKey;

    keyCallback(signup, privateKey, publicKey);
}

const inputsDisabled = (disabled: boolean) => {
    document.querySelectorAll("fieldset").forEach(
        (value) => {
       value.disabled = disabled;
    });
}

/**
 * Initiates the process of creating an account with an email.
 */
const emailSignup = async () => {
    inputsDisabled(true);
    let emailInput = document.getElementById("email") as HTMLInputElement;
    let passwordInput = document.getElementById("password") as HTMLInputElement;
    let confirmPasswordInput = document.getElementById("confirm-password") as HTMLInputElement;

    let password = passwordInput.value;
    let confirmPassword = confirmPasswordInput.value;

    if (emailInput.value && passwordIsValid(password, confirmPassword)) {
        await generateKeys(
            emailInput.value,
            passwordInput.value,
            async (signup, privKey, pubKey) => {
                const dbModule = await import("./db.js");
                let db = new dbModule.YeetFileDB();

                await db.insertVaultKeyPair(privKey, pubKey, "", success => {
                    if (success) {
                        submitSignupForm(emailInput.value, signup);
                    } else {
                        alert("Failed to insert vault key pair into indexeddb");
                        inputsDisabled(false);
                    }
                });
            });
    } else {
        inputsDisabled(false);
        alert("Missing required fields");
    }
}

/**
 * Initiates the process of creating an account ID-only account (not using an email)
 */
const accountIDOnlySignup = () => {
    let passwordInput = document.getElementById("account-password") as HTMLInputElement;
    let confirmPasswordInput = document.getElementById("account-confirm-password") as HTMLInputElement;

    if (passwordIsValid(passwordInput.value, confirmPasswordInput.value)) {
        inputsDisabled(true);
        submitSignupForm("", new interfaces.Signup());
    }
}

/**
 * submitSignupForm submits the necessary info to create a new YeetFile account
 * @param email {string}
 * @param userKeys {interfaces.Signup}
 */
const submitSignupForm = (
    email: string,
    userKeys: interfaces.Signup,
) => {
    clearMessages();

    let xhr = new XMLHttpRequest();
    xhr.open("POST", Endpoints.Signup.path, false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (email && email.length > 0) {
                window.location.assign(Endpoints.HTMLVerifyEmail.path + "?email=" + email);
            } else {
                let response = JSON.parse(xhr.responseText);
                let html = generateAccountIDSignupHTML(response.identifier, response.captcha);
                addVerifyHTML(html);
            }
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            inputsDisabled(false);
            showMessage("Error " + xhr.status + ": " + xhr.responseText, true);
        }
    };

    let hintInput = document.getElementById("password-hint") as HTMLInputElement;
    if (hintInput.value.length > MaxHintLen) {
        alert(`Password hint too long (max ${MaxHintLen} characters)`);
        return;
    }

    let sendData = userKeys;
    sendData.identifier = email;
    sendData.serverPassword = serverPassword.value;
    sendData.passwordHint = hintInput.value;

    xhr.send(JSON.stringify(sendData, jsonReplacer));
}

/**
 * Generates the "captcha" for verifying account ID-only signups
 * @param id {string} - the user's new account ID
 * @param img {string} - the base64 captcha image
 */
const generateAccountIDSignupHTML = (id: string, img: string) => {
    document.addEventListener("click", (event) => {
        if ((event.target as HTMLElement).id === verifyButtonID) {
            verifyAccountID(id);
        }
    });

    return `<br>
    <img src="data:image/jpeg;base64,${img}"<br>
    <p id="verify-note">
    Please enter the 6-digit code above to verify your account.
    </p>
    <input type="text" data-testid="account-code" id="account-code" name="code" placeholder="Code"><br>
    <button data-testid="${verifyButtonID}" id="${verifyButtonID}">Verify</button>
    `;
}

/**
 * Displays the user's account ID in a message below the signup view, allowing
 * them their one opportunity to copy the ID down.
 * @param id {string} - the user's new account ID
 */
const generateSuccessHTML = (id: string) => {
    document.addEventListener("click", (event) => {
        if ((event.target as HTMLElement).id === "goto-send") {
            window.location.assign(Endpoints.HTMLSend.path);
        }
    });

    return `<p>Your account ID is: <b data-testid="final-account-id">${id}</b> -- write this down!<br>
    This is what you will use to log in, and <b>will not be shown again.</b></p>
    <button data-testid="goto-send" id="goto-send">Start Yeeting</button>`
}

/**
 * Verifies an account ID-only signup using the 6-digit code the user entered.
 * @param id {string} - the new user ID
 */
const verifyAccountID = async (id: string) => {
    let button = document.getElementById(verifyButtonID) as HTMLButtonElement;
    let codeInput = document.getElementById("account-code") as HTMLInputElement;
    let passwordInput = document.getElementById("account-password") as HTMLInputElement;

    let password = passwordInput.value;
    button.disabled = true;

    await generateKeys(
        id,
        password,
        async (userKeys, privKey, pubKey) => {
            let body = new interfaces.VerifyAccount();
            body.id = id;
            body.code = codeInput.value;
            body.loginKeyHash = userKeys.loginKeyHash;
            body.publicKey = userKeys.publicKey;
            body.protectedPrivateKey = userKeys.protectedPrivateKey;
            body.protectedVaultFolderKey = userKeys.protectedVaultFolderKey;

            fetch(Endpoints.VerifyAccount.path, {
                method: "POST", body: JSON.stringify(body, jsonReplacer)
            }).then(async response => {
                if (response.ok) {
                    const dbModule = await import('./db.js');
                    let db = new dbModule.YeetFileDB();
                    await db.insertVaultKeyPair(privKey, pubKey, "", success => {
                        if (success) {
                            let html = generateSuccessHTML(id);
                            addVerifyHTML(html);
                        } else {
                            alert("Error inserting keys into indexed db!");
                        }
                    });
                } else {
                    button.disabled = false;
                    showMessage("Error " + await response.text(), true);
                }
            });
        });
}

const addVerifyHTML = html => {
    let div = document.getElementById("account-id-verify");
    div.style.display = "inherit";
    div.innerHTML = html;
}

const passwordIsValid = (password, confirm) => {
    if (!password || !confirm) {
        showMessage("You must fill out all available fields", true);
        return false;
    } else if (password !== confirm) {
        showMessage("Passwords do not match", true);
        return false;
    } else if (password.length < 8) {
        showMessage("Password must be at least 8 characters long", true);
        return false;
    } else {
        clearMessages();
        return true;
    }
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}