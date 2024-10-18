import {closeDialog} from "./dialogs.js";
import {PackagedPassEntry, PassEntry} from "../strict_interfaces.js";
import * as crypto from "../crypto.js";
import {PasswordGeneratorDialog} from "./password_generator.js";

export class VaultPassDialog {
    initState: string;
    dialog: HTMLDialogElement;

    entryName: HTMLInputElement;
    username: HTMLInputElement;
    password: HTMLInputElement;
    notes: HTMLTextAreaElement;

    errorSpan: HTMLSpanElement;

    primaryURL: HTMLInputElement;
    additionalURLs: HTMLInputElement[] = [];

    folderID: string;
    folderKey: CryptoKey;

    encFn: (key: CryptoKey, data: Uint8Array) => Promise<Uint8Array>;
    callback: (packaged: PackagedPassEntry) => void;

    itemKey: CryptoKey;

    constructor() {
        this.dialog = document.getElementById("password-dialog") as HTMLDialogElement;
        this.initState = this.dialog.innerHTML;
    }

    /**
     * Initializes the inner dialog HTML to its original state and sets up all
     * required input handlers.
     */
    #init = (name?: string, existing?: PassEntry, canEdit: boolean = true) => {
        this.additionalURLs = [];
        this.dialog.innerHTML = this.initState;

        let cancelBtn = document.getElementById("cancel-password") as HTMLButtonElement;
        cancelBtn.addEventListener("click", () => {
            closeDialog(this.dialog);
        });

        let submitBtn = document.getElementById("submit-password") as HTMLButtonElement;
        submitBtn.addEventListener("click", async () => {
            if (this.#validate()) {
                let passEntry = this.#finalize();
                if (this.itemKey) {
                    await this.#update(passEntry);
                } else {
                    await this.#submit(passEntry);
                }
            }
        });

        let generateBtn = document.getElementById("generate-password") as HTMLAnchorElement;
        generateBtn.addEventListener("click", () => {
            closeDialog(this.dialog);
            let generatorDialog = new PasswordGeneratorDialog();
            generatorDialog.show(newPassword => {
                this.dialog.showModal();
                this.password.value = newPassword;
            }, () => {
                this.dialog.showModal();
            })
        });

        if (this.itemKey) {
            submitBtn.innerText = "Update";
            cancelBtn.innerText = "Close";
        }

        this.entryName = document.getElementById("entry-name") as HTMLInputElement;
        this.username = document.getElementById("entry-username") as HTMLInputElement;
        this.password = document.getElementById("entry-password") as HTMLInputElement;
        this.primaryURL = document.getElementById("entry-url") as HTMLInputElement;
        this.notes = document.getElementById("entry-notes") as HTMLTextAreaElement;
        let inputs = [this.entryName, this.username, this.password, this.primaryURL, this.notes];

        if (existing && name) {
            this.entryName.value = name;
            this.username.value = existing.username;
            this.password.value = existing.password;
            this.primaryURL.value = existing.urls.length > 0 ? existing.urls[0] : "";
            this.notes.value = existing.notes;

            for (let i = 1; i < existing.urls.length; i++) {
                let urlInput = this.#addURLRow();
                urlInput.value = existing.urls[i];
                inputs.push(urlInput);
            }
        }

        this.errorSpan = document.getElementById("new-password-error") as HTMLSpanElement;

        let showPasswordToggle = document.getElementById("show-password") as HTMLInputElement;
        showPasswordToggle.checked = false;
        showPasswordToggle.addEventListener("change", () => {
            if (showPasswordToggle.checked) {
                this.password.type = "text";
            } else {
                this.password.type = "password";
            }
        });

        let addURLBtn = document.getElementById("add-url") as HTMLAnchorElement;
        addURLBtn.addEventListener("click", this.#addURLRow);

        if (!canEdit) {
            for (let i in inputs) {
                inputs[i].disabled = true;
            }

            addURLBtn.style.display = "none";
            submitBtn.style.display = "none";
            generateBtn.style.display = "none";
        }
    }

    /**
     * Validates the contents of the password dialog.
     */
    #validate = (): boolean => {
        if (!this.callback || (!this.itemKey && (!this.folderKey || !this.encFn))) {
            this.errorSpan.innerText = "Invalid dialog state";
            return false;
        } else if (this.entryName.value.length === 0) {
            this.errorSpan.innerText = "Name cannot be empty";
            return false;
        } else if (this.notes.value.length > 500) {
            this.errorSpan.innerText = "Notes cannot be > 500 characters";
            return false;
        }

        return true;
    }

    /**
     * Finalizes the PassEntry object created by the user.
     * @private
     */
    #finalize = (): PassEntry => {
        let passEntry = new PassEntry();
        passEntry.username = this.username.value;
        passEntry.password = this.password.value;
        passEntry.passwordHistory = [];
        passEntry.notes = this.notes.value;
        passEntry.urls = [this.primaryURL.value];
        for (let i in this.additionalURLs) {
            let url = this.additionalURLs[i].value;
            if (!url) {
                continue;
            }

            passEntry.urls.push(this.additionalURLs[i].value);
        }

        return passEntry;
    }

    /**
     * Validates and submits the values provided by the user in the dialog.
     * @param passEntry
     * @private
     */
    #submit = async (passEntry: PassEntry) => {
        let itemKey = await crypto.generateRandomKey();
        let encKey = await this.encFn(this.folderKey, itemKey);
        let importedKey = await crypto.importKey(itemKey);
        let packaged = await passEntry.pack(
            this.entryName.value,
            this.folderID,
            encKey,
            importedKey);
        this.callback(packaged);
        closeDialog(this.dialog);
    }

    /**
     * Updates an existing pass entry with new values
     * @param passEntry
     * @private
     */
    #update = async (passEntry: PassEntry) => {
        let packaged = await passEntry.pack(
            this.entryName.value,
            this.folderID,
            new Uint8Array,
            this.itemKey);
        this.callback(packaged);
        closeDialog(this.dialog);
    }

    /**
     * Adds a URL input row to the existing dialog
     */
    #addURLRow = (): HTMLInputElement => {
        let rowID = genRandomString(10);
        let tableRow = `
        <tr id="${rowID}">
            <td></td>
            <td>
                <input type="text" id="${rowID}-input" placeholder="https://...">
                <a id="remove-${rowID}" class="red-link">x</a>
            </td>
        </tr>`;

        if (this.additionalURLs.length === 0) {
            let urlRow = document.getElementById("url-row");
            urlRow.insertAdjacentHTML("afterend", tableRow);
        } else {
            let finalRow = this.additionalURLs[this.additionalURLs.length - 1].
                parentElement. // <td>
                parentElement; // <tr>
            finalRow.insertAdjacentHTML("afterend", tableRow);
        }

        let newRowInput = document.getElementById(`${rowID}-input`) as HTMLInputElement;
        this.additionalURLs.push(newRowInput);

        let removeRow = document.getElementById(`remove-${rowID}`) as HTMLAnchorElement;
        removeRow.addEventListener("click", () => {
            document.getElementById(rowID).remove();
            this.additionalURLs = this.additionalURLs.filter(el => {
                return el.id.indexOf(rowID) < 0;
            });
        });

        return newRowInput;
    }

    /**
     * Creates a new instance of the new password dialog within the current
     * vault folder context.
     * @param folderID {string} - the vault folder ID
     * @param folderKey {CryptoKey} - the vault folder key
     * @param encFn {(CryptoKey, Uint8Array) => Uint8Array} - the function for
     * encrypting the item's key (AES-GCM or RSA).
     * @param callback {(PackagedPassEntry)} - The callback for successful submissions
     */
    create = (
        folderID: string,
        folderKey: CryptoKey,
        encFn: (key: CryptoKey, data: Uint8Array) => Promise<Uint8Array>,
        callback: (packaged: PackagedPassEntry) => void,
    ) => {
        this.folderID = folderID;
        this.folderKey = folderKey;
        this.encFn = encFn;
        this.callback = callback;
        this.itemKey = undefined;

        this.#init();
        this.dialog.showModal();
    }

    /**
     * Sets up the password dialog with existing values to let the user edit them.
     * @param name
     * @param entry
     * @param key
     * @param callback
     * @param canEdit
     */
    edit = (
        name: string,
        entry: PassEntry,
        key: CryptoKey,
        callback: (packaged: PackagedPassEntry) => void,
        canEdit: boolean,
    ) => {
        this.itemKey = key;
        this.callback = callback;

        this.#init(name, entry, canEdit);
        this.dialog.showModal();
    }
}