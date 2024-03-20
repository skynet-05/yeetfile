const login = async () => {
    let identifier = document.getElementById("identifier");
    let password = document.getElementById("password");

    if (!isValidIdentifier(identifier.value)) {
        return;
    }

    let userKey = await generateUserKey(identifier.value, password.value);
    let loginKeyHash = await generateLoginKeyHash(userKey, password.value);

    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/login", false);
    xhr.setRequestHeader("Content-Type", "application/json");

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            window.location = "/";
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            showErrorMessage("Error " + xhr.status + ": " + xhr.responseText);
        }
    };

    xhr.send(JSON.stringify({
        identifier: identifier.value,
        loginKeyHash: Array.from(loginKeyHash)
    }));
}

const isValidIdentifier = (identifier) => {
    if (identifier.includes("@")) {
        return true;
    } else {
        if (identifier.length !== 16) {
            showErrorMessage("Invalid email or 16-digit account ID");
            return false;
        }

        return true;
    }
}