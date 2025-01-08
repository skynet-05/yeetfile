import {Endpoints} from "./endpoints.js";

let selectedSendUpgrade: string;
let selectedVaultUpgrade: string;

const init = () => {
    setupQuantityListeners();
    setupUpgradeButtons();
    setupBTCPayToggle();
    setupCheckoutButton();
}

const updateCheckoutButton = () => {
    let checkoutBtn = document.getElementById("checkout-btn") as HTMLButtonElement;
    checkoutBtn.disabled = !(selectedSendUpgrade || selectedVaultUpgrade);
}

const setupQuantityListeners = () => {
    document.querySelectorAll(".pricing-box").forEach(container => {
        const input = container.querySelector(".quantity-input") as HTMLInputElement;
        const links = container.querySelectorAll("a") as NodeListOf<HTMLAnchorElement>;

        // Listen for changes to the quantity input
        input.addEventListener("input", () => {
            const value = input.value;

            links.forEach(link => {
                const baseUrl = link.href.split("&")[0];
                link.href = `${baseUrl}&quantity=${encodeURIComponent(value)}`;
            });
        });
    });
}

const setupUpgradeButtons = () => {
    document.querySelectorAll(".select-btn").forEach(container => {
        let element = container as HTMLButtonElement;

        element.addEventListener("click", e => {
            let id = container.id;
            let upgradeType = element.dataset.upgradeType;
            let deselect = false;

            if (upgradeType === "send") {
                if (selectedSendUpgrade === id) {
                    deselect = true;
                    selectedSendUpgrade = "";
                } else {
                    selectedSendUpgrade = id;
                }
            } else {
                if (selectedVaultUpgrade === id) {
                    deselect = true;
                    selectedVaultUpgrade = "";
                } else {
                    selectedVaultUpgrade = id;
                }
            }

            document.querySelectorAll(".pricing-box").forEach(div => {
                div.querySelectorAll("button").forEach(btn => {
                    let htmlBtn = btn as HTMLButtonElement;
                    if (htmlBtn.dataset.upgradeType === upgradeType) {
                        htmlBtn.disabled = !deselect && htmlBtn.id !== id;
                    }
                });
            });

            if (deselect) {
                element.classList.add("accent-btn");
                element.classList.remove("destructive-btn");
                element.innerText = "Select";
            } else {
                element.classList.remove("accent-btn");
                element.classList.add("destructive-btn");
                element.innerText = "Remove";
            }

            updateCheckoutButton();
        });
    });
}

const setupCheckoutButton = () => {
    let checkoutBtn = document.getElementById("checkout-btn") as HTMLButtonElement;
    checkoutBtn.addEventListener("click", () => {
        if (!selectedSendUpgrade && !selectedVaultUpgrade) {
            return;
        }

        let vaultQuantity = "1";
        if (selectedVaultUpgrade) {
            let quantityID = `${selectedVaultUpgrade}-quantity`;
            let quantityEl = document.getElementById(quantityID) as HTMLInputElement;
            vaultQuantity = quantityEl.value;

            let quantityInt = parseInt(vaultQuantity);
            if (isNaN(quantityInt) || quantityInt < 0 || quantityInt > 12) {
                vaultQuantity = "1";
            }
        }

        let link = new URL(window.location.origin + Endpoints.StripeCheckout.path);
        if (selectedSendUpgrade) {
            link.searchParams.set("send-upgrade", selectedSendUpgrade);
        }

        if (selectedVaultUpgrade) {
            link.searchParams.set("vault-upgrade", selectedVaultUpgrade);
            link.searchParams.set("vault-quantity", vaultQuantity);
        }

        checkoutBtn.innerText = "Processing...";
        checkoutBtn.disabled = true;

        window.location.assign(link);
    });
}

const setupBTCPayToggle = () => {
    let cb = document.getElementById("btcpay-cb") as HTMLInputElement;
    let url = new URL(window.location.href);

    cb.addEventListener("click", () => {
        url.searchParams.set("btcpay", cb.checked ? "1" : "0");
        window.location.assign(url);
    })
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}