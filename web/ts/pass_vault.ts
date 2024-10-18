import {VaultView, VaultViewType, prep} from "./vault.js";
import {YeetFileDB} from "./db.js";
import {Endpoints} from "./endpoints.js";

const longWordlistFile = "eff_long_wordlist.json";
const shortWordlistFile = "eff_short_wordlist.json";
const longWordlistURL = Endpoints.format(
    Endpoints.StaticFile,
    "json",
    longWordlistFile);
const shortWordlistURL = Endpoints.format(
    Endpoints.StaticFile,
    "json",
    shortWordlistFile);

const db = new YeetFileDB();

const init = () => {
    prep((privKey, pubKey) => {
        loadVaultView(privKey, pubKey);
    });

    loadWordLists();
}

const loadWordLists = () => {
    db.fetchWordlists(success => {
        if (!success) {
            fetchWordlists();
        }
    });
}

const fetchWordlists = () => {
    fetch(longWordlistURL).then(response => {
        return response.json()
    }).then(longWordlist => {
        fetch(shortWordlistURL).then(response => {
            return response.json();
        }).then(shortWordlist => {
            db.storeWordlists(longWordlist, shortWordlist, success => {
                if (!success) {
                    alert("Failed to store passphrase wordlists -- " +
                        "passphrase generation may not work");
                    console.error("Failed to insert wordlists");
                }
            });
        }).catch(() => {
            console.error("Error fetching short wordlist");
        });
    }).catch(() => {
        console.error("Error fetching long wordlist");
    });
}

const loadVaultView = (privKey: CryptoKey, pubKey: CryptoKey) => {
    let vaultView = new VaultView(VaultViewType.PassVault, privKey, pubKey);
    vaultView.initialize();
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}