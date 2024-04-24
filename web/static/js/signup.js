import * as crypto from "./crypto.js";

document.addEventListener("DOMContentLoaded", () => {
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
});

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

const emailSignup = async (btn) => {
    let emailInput = document.getElementById("email");
    let passwordInput = document.getElementById("password");
    let confirmPasswordInput = document.getElementById("confirm-password");

    if (emailInput.value && passwordIsValid(passwordInput.value, confirmPasswordInput.value)) {
        passwordInput.disabled = true;
        confirmPasswordInput.disabled = true;

        let userKey = await crypto.generateUserKey(emailInput.value, passwordInput.value);
        let loginKeyHash = await crypto.generateLoginKeyHash(userKey, passwordInput.value);
        let keyPair = await crypto.generateKeyPair();
        let publicKey = await crypto.exportKey(keyPair.publicKey, "spki");
        let privateKey = await crypto.exportKey(keyPair.privateKey, "pkcs8");
        let protectedKey = await crypto.encryptChunk(userKey, privateKey);
        let folderKey = await crypto.generateRandomKey();
        let protectedRootFolderKey = await crypto.encryptChunk(keyPair.publicKey, folderKey);

        // Store in indexeddb
        let encodedKey = await arrayToBase64(protectedKey);

        crypto.ingestProtectedKey(userKey, encodedKey, () => {
            new YeetFileDB().insertVaultKeyPair(keyPair.privateKey, keyPair.publicKey);
            submitSignupForm(btn, emailInput.value, loginKeyHash, publicKey, protectedKey, protectedRootFolderKey);
        });
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

const submitSignupForm = (submitBtn, email, loginKeyHash, publicKey, protectedKey, rootFolderKey) => {
    submitBtn.disabled = true;

    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/signup", false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (email && email.length > 0) {
                window.location = "/verify?email=" + email;
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

    xhr.send(JSON.stringify({
        identifier: email ? email : "",
        loginKeyHash: loginKeyHash ? Array.from(loginKeyHash) : loginKeyHash,
        protectedKey: protectedKey ? Array.from(protectedKey) : protectedKey,
        publicKey: publicKey ? Array.from(publicKey) : publicKey,
        rootFolderKey: rootFolderKey ? Array.from(rootFolderKey) : rootFolderKey,
    }));
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

    let userKey = await crypto.generateUserKey(id, password);
    let loginKeyHash = await crypto.generateLoginKeyHash(userKey, password);
    let keyPair = await crypto.generateKeyPair();
    let publicKey = await crypto.exportKey(keyPair.publicKey, "spki");
    let privateKey = await crypto.exportKey(keyPair.privateKey, "pkcs8");
    let protectedKey = await crypto.encryptChunk(userKey, privateKey);
    let folderKey = await crypto.generateRandomKey();
    let protectedRootFolderKey = await crypto.encryptRSA(keyPair.publicKey, folderKey);

    let encodedKey = await arrayToBase64(protectedKey);

    crypto.ingestProtectedKey(userKey, encodedKey, () => {
        new YeetFileDB().insertVaultKeyPair(keyPair.privateKey, keyPair.publicKey);

        let xhr = new XMLHttpRequest();
        xhr.open("POST", "/verify-account", false);
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
            loginKeyHash: Array.from(loginKeyHash),
            protectedKey: Array.from(protectedKey),
            publicKey: Array.from(publicKey),
            rootFolderKey: Array.from(protectedRootFolderKey)
        }));
    });
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