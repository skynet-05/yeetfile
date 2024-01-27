document.addEventListener("DOMContentLoaded", () => {
    setupTypeToggles();

    // Email signup fields
    let emailInput = document.getElementById("email");
    let passwordInput = document.getElementById("password");
    let confirmPasswordInput = document.getElementById("confirm-password");
    let emailSignupButton = document.getElementById("create-email-account");

    emailSignupButton.addEventListener("click", (event) => {
        event.preventDefault();

        if (!emailInput.value || !passwordInput || !confirmPasswordInput) {
            addAuthError("You must fill out all available fields");
            return;
        } else if (passwordInput.value !== confirmPasswordInput.value) {
            addAuthError("Passwords do not match");
            return;
        }

        submitSignupForm(emailInput.value, passwordInput.value, emailSignupButton);
    });

    // Account ID only fields
    let accountSignupButton = document.getElementById("create-id-only-account");

    accountSignupButton.addEventListener("click", (event) => {
        event.preventDefault();

        submitSignupForm("", "", accountSignupButton);
    })
});

const submitSignupForm = (email, password, submitBtn) => {
    submitBtn.disabled = true;

    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/signup", false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            if (email.length > 0) {
                window.location = "/verify?email=" + email;
            } else {
                let html = generateAccountIDSignupHTML(xhr.responseText);
                addAuthHTML(html);
            }
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            submitBtn.disabled = false;
            addAuthError("Error " + xhr.status + ": " + xhr.responseText);
        }
    };

    xhr.send(JSON.stringify({
        email: email,
        password: password
    }));
}

const generateAccountIDSignupHTML = (id) => {
    return `
    <p>Your account ID is: <b>${id}</b> -- write this down!
    This is what you will use to log in.</p>
    <button onclick="goToAccount()">Go To Account</button>
    `;
}

const goToAccount = () => {
    window.location = "/account";
}