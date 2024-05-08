document.addEventListener("DOMContentLoaded", () => {
    let logoutBtn = document.getElementById("logout-btn");
    logoutBtn.addEventListener("click", logout);

    let recyclePaymentIDBtn = document.getElementById("recycle-payment-id");
    recyclePaymentIDBtn.addEventListener("click", recyclePaymentID);
});

const logout = () => {
    let confirmMsg = "Log out of YeetFile?";
    if (confirm(confirmMsg)) {
        new YeetFileDB().removeKeys(success => {
            if (success) {
                window.location = "/logout";
            }
        });
    }
}

const recyclePaymentID = () => {
    let confirmMsg = "Are you sure you want to recycle your payment ID? " +
        "This will remove all records of past payments you've made.";
    if (confirm(confirmMsg)) {
        fetch("/api/recycle-payment-id").then(() => {
            window.location = "/account";
        }).catch(() => {
            alert("Error recycling payment id");
        });
    }
}