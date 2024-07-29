import * as crypto from "./crypto.js";
import * as endpoints from "./endpoints.js";
import {YeetFileDB} from "./db.js";

const init = () => {
    setupToggles();

    // Email signup
    let emailSignupButton = document.getElementById("create-email-account");
    emailSignupButton.addEventListener("click", async (event) => {
        event.preventDefault();
        await emailSignup(emailSignupButton);
    });

    // Account ID only signup
    let accountSignupButton = document.getElementById("create-id-only-account");
    accountSignupButton.addEventListener("click", (event) => {
        event.preventDefault();
        accountIDOnlySignup(accountSignupButton);
    })
};

const setupToggles = () => {
    let emailToggle = document.getElementById("email-signup");
    let emailDiv = document.getElementById("email-div");

    let idToggle = document.getElementById("id-signup");
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
 * @returns {Promise<{
 *     loginKeyHash: Uint8Array,
 *     protectedKey: Uint8Array,
 *     privateKey: Uint8Array,
 *     publicKey: Uint8Array,
 *     rootFolderKey: Uint8Array
 * }>}
 */
const generateKeys = async (identifier, password) => {
    let userKey = await crypto.generateUserKey(identifier, password);
    let loginKeyHash = await crypto.generateLoginKeyHash(userKey, password);
    let keyPair = await crypto.generateKeyPair();
    let publicKey = await crypto.exportKey(keyPair.publicKey, "spki");
    let privateKey = await crypto.exportKey(keyPair.privateKey, "pkcs8");
    let protectedKey = await crypto.encryptChunk(userKey, privateKey);
    let folderKey = await crypto.generateRandomKey();
    let protectedRootFolderKey = await crypto.encryptRSA(keyPair.publicKey, folderKey);

    return {
        "loginKeyHash": loginKeyHash,
        "publicKey": publicKey,
        "privateKey": privateKey,
        "protectedKey": protectedKey,
        "rootFolderKey": protectedRootFolderKey
    }
}

const emailSignup = async (btn) => {
    let emailInput = document.getElementById("email");
    let passwordInput = document.getElementById("password");
    let confirmPasswordInput = document.getElementById("confirm-password");

    if (emailInput.value && passwordIsValid(passwordInput.value, confirmPasswordInput.value)) {
        passwordInput.disabled = true;
        confirmPasswordInput.disabled = true;

        let userKeys = await generateKeys(emailInput.value, passwordInput.value);

        await new YeetFileDB().insertVaultKeyPair(userKeys["privateKey"], userKeys["publicKey"], "");
        submitSignupForm(btn, emailInput.value, userKeys);
    }
}

const accountIDOnlySignup = (btn) => {
    let passwordInput = document.getElementById("account-password");
    let confirmPasswordInput = document.getElementById("account-confirm-password");

    if (passwordIsValid(passwordInput.value, confirmPasswordInput.value)) {
        passwordInput.disabled = true;
        confirmPasswordInput.disabled = true;
        submitSignupForm(btn);
    }
}

/**
 * submitSignupForm submits the necessary info to create a new YeetFile account
 * @param submitBtn {}
 * @param email {string}
 * @param userKeys {{
 *     loginKeyHash: Uint8Array,
 *     protectedKey: Uint8Array,
 *     privateKey: Uint8Array,
 *     publicKey: Uint8Array,
 *     rootFolderKey: Uint8Array
 * }}
 */
const submitSignupForm = (submitBtn, email, userKeys) => {
    submitBtn.disabled = true;

    let xhr = new XMLHttpRequest();
    xhr.open("POST", endpoints.Signup, false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (email && email.length > 0) {
                window.location = "/verify-email?email=" + email;
            } else {
                let response = JSON.parse(xhr.responseText);
                let html = generateAccountIDSignupHTML(response.identifier, response.captcha);
                addVerifyHTML(html);
            }
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            submitBtn.disabled = false;
            showErrorMessage("Error " + xhr.status + ": " + xhr.responseText);
        }
    };

    let sendData = {identifier: email ? email : ""};
    if (userKeys) {
        sendData = {
            ...sendData,
            loginKeyHash: Array.from(userKeys["loginKeyHash"]),
            protectedKey: Array.from(userKeys["protectedKey"]),
            publicKey: Array.from(userKeys["publicKey"]),
            rootFolderKey: Array.from(userKeys["rootFolderKey"]),
        }
    }

    xhr.send(JSON.stringify(sendData));
}

const generateAccountIDSignupHTML = (id, img) => {
    document.addEventListener("click", (event) => {
        if (event.target.id === "verify-account") {
            verifyAccountID(id);
        }
    });

    return `
    <img src="data:image/jpeg;base64,${img}"<br>
    <p>Please enter the 6-digit code above to verify your account.</p>
    <input type="text" id="account-code" name="code" placeholder="Code"><br>
    <button id="verify-account">Verify</button>
    `;
}

const generateSuccessHTML = (id) => {
    document.addEventListener("click", (event) => {
        if (event.target.id === "goto-account") {
            window.location = "/account";
        }
    });

    return `<p>Your account ID is: <b>${id}</b> -- write this down!
    This is what you will use to log in, and will not be shown again.</p>
    <button id="goto-account">Go To Account</button>`
}

const verifyAccountID = async id => {
    let button = document.getElementById("verify-account");
    let codeInput = document.getElementById("account-code");

    let password = document.getElementById("account-password").value;

    button.disabled = true;

    let userKeys = await generateKeys(id, password);
    await new YeetFileDB().insertVaultKeyPair(userKeys["privateKey"], userKeys["publicKey"], "");

    let xhr = new XMLHttpRequest();
    xhr.open("POST", endpoints.VerifyAccount, false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            let html = generateSuccessHTML(id);
            addVerifyHTML(html);
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            button.disabled = false;
            showErrorMessage("Error " + xhr.status + ": " + xhr.responseText);
        }
    };

    xhr.send(JSON.stringify({
        id: id,
        code: codeInput.value,
        loginKeyHash: Array.from(userKeys["loginKeyHash"]),
        protectedKey: Array.from(userKeys["protectedKey"]),
        publicKey: Array.from(userKeys["publicKey"]),
        rootFolderKey: Array.from(userKeys["rootFolderKey"])
    }));
}

const addVerifyHTML = html => {
    let div = document.getElementById("account-id-verify");
    div.style.display = "inherit";
    div.innerHTML = html;
}

const passwordIsValid = (password, confirm) => {
    if (!password || !confirm) {
        showErrorMessage("You must fill out all available fields");
        return false;
    } else if (password !== confirm) {
        showErrorMessage("Passwords do not match");
        return false;
    } else if (password.length < 7) {
        showErrorMessage("Password must be at least 7 characters long");
        return false;
    } else {
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