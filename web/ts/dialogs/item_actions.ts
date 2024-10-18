import {VaultViewFolder, VaultViewItem} from "../vault.js";
import * as dialogs from "./dialogs.js";
import {DialogSignal} from "./dialogs.js";

export class ActionsDialog {
    dialog: HTMLDialogElement;
    initState: string;
    callback: (item: VaultViewItem|VaultViewFolder, s: DialogSignal) => void;
    
    constructor(callback: (item: VaultViewItem|VaultViewFolder, s: DialogSignal) => void) {
        this.dialog = document.getElementById("actions-dialog") as HTMLDialogElement;
        this.initState = this.dialog.innerHTML;
        this.callback = callback;
    }

    #init = () => {
        this.dialog = document.getElementById("actions-dialog") as HTMLDialogElement;
        if (!this.initState) {
            this.initState = this.dialog.innerHTML;
        }

        this.dialog.innerHTML = this.initState;
    }
    
    show = (item: VaultViewItem|VaultViewFolder) => {
        this.#init()

        let isFolder = item instanceof VaultViewFolder;
        let title = document.getElementById("actions-dialog-title") as HTMLHeadingElement;
        if (isFolder) {
            title.innerText = "Folder: " + item.decName;
        } else {
            title.innerText = item.decName;
        }

        let actionDownload = document.getElementById("action-download");
        if (!isFolder && (!item.passwordData || item.passwordData.length === 0)) {
            actionDownload.style.display = "flex";
            actionDownload.addEventListener("click", event => {
                event.stopPropagation();
                dialogs.closeDialog(this.dialog);
                this.callback(item, dialogs.DialogSignal.Download);
            });
        } else {
            actionDownload.style.display = "none";
        }

        let actionRename = document.getElementById("action-rename");
        if (item.canModify) {
            actionRename.style.display = "flex";
            actionRename.addEventListener("click", (event) => {
                event.stopPropagation();
                dialogs.closeDialog(this.dialog);
                this.callback(item, dialogs.DialogSignal.Rename);
            });
        } else {
            actionRename.style.display = "none";
        }

        let actionShare = document.getElementById("action-share");
        if (item.isOwner) {
            actionShare.style.display = "flex";
            actionShare.addEventListener("click", async event => {
                event.stopPropagation();
                this.callback(item, dialogs.DialogSignal.Share.valueOf());
                dialogs.closeDialog(this.dialog);
            });
        } else {
            actionShare.style.display = "none";
        }

        let actionDelete = document.getElementById("action-delete");
        if (item.isOwner || item.canModify) {
            actionDelete.style.display = "flex";
            actionDelete.addEventListener("click", event => {
                event.stopPropagation();
                this.callback(item, dialogs.DialogSignal.Delete);
                dialogs.closeDialog(this.dialog);
            });
        } else {
            actionDelete.style.display = "none";
        }

        let actionRemove = document.getElementById("action-remove");
        if (!item.isOwner) {
            actionRemove.style.display = "flex";
            actionRemove.addEventListener("click", event => {
                event.stopPropagation();
                this.callback(item, dialogs.DialogSignal.Remove);
                dialogs.closeDialog(this.dialog);
            });
        } else {
            actionRemove.style.display = "none";
        }

        let cancel = document.getElementById("cancel-action");
        cancel.addEventListener("click", () => dialogs.closeDialog(this.dialog));
        this.dialog.showModal();
    }
}