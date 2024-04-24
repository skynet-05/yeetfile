document.addEventListener("DOMContentLoaded", () => {
    let logoutBtn = document.getElementById("logout-btn");
    logoutBtn.addEventListener("click", logout);
});

const logout = () => {
    new YeetFileDB().removeKeys(success => {
        if (success) {
            window.location = "/logout";
        }
    });

}