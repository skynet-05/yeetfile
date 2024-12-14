import { VaultView, VaultViewType, prep } from "./vault.js";

const init = () => {
    prep((privKey, pubKey) => {
        loadVaultView(privKey, pubKey);
    });
}

const loadVaultView = (privKey: CryptoKey, pubKey: CryptoKey) => {
    let vaultView = new VaultView(VaultViewType.FileVault, privKey, pubKey);
    vaultView.initialize();
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}
