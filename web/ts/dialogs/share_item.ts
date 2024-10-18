import * as transfer from "../transfer.js";
import {closeDialog, DialogSignal} from "./dialogs.js";

export class ShareContentDialog {
    dialog: HTMLDialogElement;
    target: HTMLInputElement;
    modify: HTMLInputElement;

    submit: HTMLButtonElement;
    cancel: HTMLButtonElement;

    loading: HTMLElement;
    table: HTMLTableElement;
    tableBody: HTMLTableElement;

    constructor() {
        this.dialog = document.getElementById("share-dialog") as HTMLDialogElement;
        this.target = document.getElementById("share-target") as HTMLInputElement;
        this.modify = document.getElementById("share-modify") as HTMLInputElement;

        this.submit = document.getElementById("submit-share") as HTMLButtonElement;
        this.cancel = document.getElementById("cancel-share") as HTMLButtonElement;

        this.loading = document.getElementById("share-loading");
        this.table = document.getElementById("share-table") as HTMLTableElement;
        this.tableBody = document.getElementById("share-table-body") as HTMLTableElement;
    }

    /**
     * Display the dialog for sharing/modifying shares of a file or folder
     * @param id {string} - The file or folder ID
     * @param rawKey {ArrayBuffer} - The unencrypted key for the item
     * @param isFolder {boolean} - True if the item is a folder
     * @param callback {function(DialogSignal)} - Callback indicating the action performed
     */
    show = (
        id: string,
        rawKey: ArrayBuffer,
        isFolder: boolean,
        callback: (s: DialogSignal) => void,
    ) => {
        this.loading.style.display = "inherit";
        this.table.style.display = "none";
        this.tableBody.innerHTML = "";

        transfer.getSharedUsers(id, isFolder).then(response => {
            this.loading.style.display = "none";
            if (response && (response as Array<any>).length !== 0) {
                this.table.style.display = "inherit";
            } else {
                return;
            }

            for (let i = 0; i < (response as Array<any>).length; i++) {
                generateShareRow(id, this.tableBody, response[i], isFolder, callback);
            }
        });

        this.submit.addEventListener("click", async event => {
            event.stopPropagation();

            if (this.target.value.length === 0) {
                alert("Must enter a YeetFile user's email or account ID for sharing");
                return;
            }

            updateButton(this.submit, true, "Sharing...");

            let target = this.target.value;
            let canModify = this.modify.checked;
            transfer.shareItem(target, rawKey, id, canModify, isFolder).then(response => {
                let name = target;
                if (!target.includes("@")) {
                    name = "*" + name.substring(name.length - 4, name.length);
                }
                generateShareRow(
                    id,
                    this.tableBody,
                    {id: response.id, recipientName: name},
                    isFolder,
                    callback);
                updateButton(this.submit, false, "Share");
                callback(DialogSignal.Share);
            }).catch(() => {
                updateButton(this.submit, false, "Share");
            });
        });

        this.cancel.addEventListener("click", event => {
            event.stopPropagation();
            closeDialog(this.dialog);
            callback(DialogSignal.Cancel);
        })

        this.dialog.showModal();
    }
}

const generateShareRow = (id, tableBody, recipient, isFolder, callback) => {
    let row = `<tr id="share-${recipient.id}">
<td>${recipient.recipientName}</td>
<td><input id="can-modify-${recipient.id}" type="checkbox" ${recipient.canModify ? "checked" : ""}></td>
<td><img id="remove-share-${recipient.id}" class="small-icon red-icon" src="/static/icons/remove.svg"></td>
</tr>`;

    tableBody.parentElement.style.display = "table";

    tableBody.innerHTML += row;
    tableBody.addEventListener("click", event => {
        if (event.target.id === `can-modify-${recipient.id}`) {
            let cb = document.getElementById(`can-modify-${recipient.id}`) as HTMLInputElement;
            transfer.changeSharedItemPerms(id, recipient.id, cb.checked, isFolder);
        } else if (event.target.id === `remove-share-${recipient.id}`) {
            if (confirm(`Remove user '${recipient.recipientName}' from shared content?`)) {
                transfer.removeUserFromShared(id, recipient.id, isFolder).then(() => {
                    callback(DialogSignal.Share);

                    document.getElementById(`share-${recipient.id}`).remove();
                    // Hide table if there aren't any users left
                    if (tableBody.children.length === 0) {
                        tableBody.parentElement.style.display = "none";
                    }
                });
            }

        }
    });
}