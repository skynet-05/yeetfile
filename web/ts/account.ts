import {Endpoints} from "./endpoints.js";
import {YeetFileDB} from "./db.js";

const init = () => {
    let logoutBtn = document.getElementById("logout-btn");
    logoutBtn.addEventListener("click", logout);

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

if (document.readyState !== 'loading') {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}