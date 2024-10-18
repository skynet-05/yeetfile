import {YeetFileDB} from "../db";

export class ProtectedVaultDialog {
    dialog: HTMLDialogElement;
    input: HTMLInputElement;
    error: HTMLSpanElement;
    cancel: HTMLButtonElement;
    submit: HTMLButtonElement;

    constructor() {
        this.dialog = document.getElementById("vault-pass-dialog") as HTMLDialogElement;
        this.input = document.getElementById("vault-pass") as HTMLInputElement;
        this.error = document.getElementById("pass-error") as HTMLSpanElement;
        this.cancel = document.getElementById("cancel-pass") as HTMLButtonElement;
        this.submit = document.getElementById("submit-pass") as HTMLButtonElement;
    }

    /**
     * Display a dialog for the current vault password (if one was set when logging in)
     * @param yeetfileDB {YeetFileDB} - The yeetfile indexeddb instance
     * @param callback {function(CryptoKey, CryptoKey)}
     * @param errorMsg {string|null}
     */
    show = (
        yeetfileDB: YeetFileDB,
        callback: (privKey: CryptoKey, pubKey: CryptoKey) => void,
        errorMsg: string|null,
    ) => {
        this.input.value = "";

        if (errorMsg) {
            this.error.innerText = errorMsg;
        }

        this.cancel.addEventListener("click", () => {
            this.dialog.close();
            window.location.assign("/");
        }, {once: true});

        this.submit.addEventListener("click", async () => {
            let password = this.input.value;
            this.dialog.close();
            yeetfileDB.getVaultKeyPair(password, false).then(([privKey, pubKey]) => {
                callback(privKey as CryptoKey, pubKey as CryptoKey);
            }).catch(e => {
                this.show(yeetfileDB, callback, e);
                return;
            });
        }, {once: true});

        this.dialog.showModal();
    }
}