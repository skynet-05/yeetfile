import {Endpoints} from "./endpoints.js";
import {YeetFileDB} from "./db.js";
import * as endpoints from "./endpoints.js";
import * as interfaces from "./interfaces.js";

const init = () => {
    let logoutBtn = document.getElementById("logout-btn");
    logoutBtn.addEventListener("click", logout);

    let deleteBtn = document.getElementById("delete-btn");
    deleteBtn.addEventListener("click", deleteAccount);

    let changePwBtn = document.getElementById("change-pw-btn");
    changePwBtn.addEventListener("click", changePassword);

    // let recyclePaymentIDBtn = document.getElementById("recycle-payment-id");
    // recyclePaymentIDBtn.addEventListener("click", recyclePaymentID);

    let yearlyToggle = document.getElementById("yearly-toggle");
    yearlyToggle.addEventListener("click", () => { window.location.assign("/account?yearly=1") });

    let monthlyToggle = document.getElementById("monthly-toggle");
    monthlyToggle.addEventListener("click", () => { window.location.assign("/account") });
}

const logout = () => {
    let confirmMsg = "Log out of YeetFile?";
    if (confirm(confirmMsg)) {
        new YeetFileDB().removeKeys(success => {
            if (success) {
                window.location.assign(Endpoints.Logout.path);
            } else {
                alert("Error removing keys");
            }
        });
    }
}

const deleteAccount = () => {
    let confirmMsg = "Are you sure you want to delete your account? This can " +
        "not be undone."
    if (!confirm(confirmMsg)) {
        return;
    }

    let promptMsg = "Enter your login email or account ID -- below to delete " +
        "your account."

    let id = prompt(promptMsg);
    if (id.length > 0) {
        let request = new interfaces.DeleteAccount();
        request.identifier = id;

        fetch(Endpoints.Account.path, {
            method: "DELETE",
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(request)
        }).then(async response => {
            if (response.ok) {
                alert("Your account has been permanently yeeted.");
                window.location.assign("/");
            } else {
                let errMsg = await response.text()
                alert("There was an error deleting your account: " + errMsg);
                return;
            }
        }).catch(error => {
            console.error(error);
            alert("There was an error deleting your account!");
        })
    }
}

const recyclePaymentID = () => {
    let confirmMsg = "Are you sure you want to recycle your payment ID? " +
        "This will remove all records of past payments you've made.";
    if (confirm(confirmMsg)) {
        fetch("/api/recycle-payment-id").then(() => {
            window.location.assign("/account");
        }).catch(() => {
            alert("Error recycling payment id");
        });
    }
}

const changePassword = () => {
    window.location.assign(Endpoints.HTMLChangePassword.path);
}

if (document.readyState !== 'loading') {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}