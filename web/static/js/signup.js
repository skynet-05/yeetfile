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

        let userKey = await generateUserKey(emailInput.value, passwordInput.value);
        let loginKeyHash = await generateLoginKeyHash(userKey, passwordInput.value);
        let storageKey = await generateStorageKey();
        let protectedKey = await encryptChunk(userKey, storageKey);

        submitSignupForm(emailInput.value, loginKeyHash, protectedKey, btn);
    }
}

const accountIDOnlySignup = (btn) => {
    let passwordInput = document.getElementById("account-password");
    let confirmPasswordInput = document.getElementById("account-confirm-password");

    if (passwordIsValid(passwordInput.value, confirmPasswordInput.value)) {
        passwordInput.disabled = true;
        confirmPasswordInput.disabled = true;
        submitSignupForm("", undefined, undefined, btn);
    }
}

const submitSignupForm = (email, loginKeyHash, protectedKey, submitBtn) => {
    submitBtn.disabled = true;

    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/signup", false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (email.length > 0) {
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
        identifier: email,
        loginKeyHash: loginKeyHash ? Array.from(loginKeyHash) : loginKeyHash,
        protectedKey: protectedKey ? Array.from(protectedKey) : protectedKey
    }));
}

const generateAccountIDSignupHTML = (id, img) => {
    return `
    <img src="data:image/jpeg;base64,${img}"<br>
    <p>Please enter the 6-digit code above to verify your account.</p>
    <input type="text" id="account-code" name="code" placeholder="Code"><br>
    <button id="verify-account" onclick="verifyAccountID('${id}')">Verify</button>
    `;
}

const generateSuccessHTML = (id) => {
    return `<p>Your account ID is: <b>${id}</b> -- write this down!
    This is what you will use to log in, and will not be shown again.</p>
    <button onclick="goToAccount()">Go To Account</button>`
}

const goToAccount = () => {
    window.location = "/account";
}

const verifyAccountID = async id => {
    let button = document.getElementById("verify-account");
    let codeInput = document.getElementById("account-code");

    let password = document.getElementById("account-password").value;

    button.disabled = true;

    let userKey = await generateUserKey(id, password);
    let loginKeyHash = await generateLoginKeyHash(userKey, password);
    let storageKey = await generateStorageKey();
    let protectedKey = await encryptChunk(userKey, storageKey);

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