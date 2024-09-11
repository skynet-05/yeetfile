import * as crypto from "./crypto.js";
import * as dialogs from "./dialogs.js";
import * as transfer from "./transfer.js";
import * as constants from "./constants.js";
import { Endpoints } from "./endpoints.js";
import * as interfaces from "./interfaces.js"
import { YeetFileDB } from "./db.js";

const gapFill = 9;
const actionIDPrefix = "action";
const folderIDPrefix = "load-folder";
const itemIDPrefix = "load-item";
const sharedWithSuffix = "sharedwith";

const emptyRow = `<tr class="blank-row"><td colspan="4"></td></tr>`
const vaultHome = `<a id="${folderIDPrefix}-" href="#">Home</a>`;
const folderPlaceholder = `<a href="/vault">← Back</a> / ...`

let folderDialog;
let subfolderParentID;
let privateKey;
let publicKey;
let folderID = "";
let folderKey;

let pauseInteractions = false;

let vaultItems = {};
let vaultFolders = {};

const init = () => {
    folderID = getFolderID();
    if (folderID.length > 0) {
        let vaultStatus = document.getElementById("vault-status");
        vaultStatus.innerHTML = folderPlaceholder;
    }

    let yeetfileDB = new YeetFileDB();
    yeetfileDB.isPasswordProtected(isProtected => {
        if (isProtected) {
            showVaultPassDialog(yeetfileDB);
        } else {
            yeetfileDB.getVaultKeyPair("", (privKey, pubKey) => {
                privateKey = privKey;
                publicKey = pubKey;

                loadFolder(folderID);
            }, () => {
                alert("Failed to decrypt vault keys");
            });
        }
    });

    let vaultUploadBtn = document.getElementById("vault-upload") as HTMLButtonElement;
    let vaultFileInput = document.getElementById("file-input") as HTMLInputElement;
    vaultUploadBtn.addEventListener("click", () => {
        vaultFileInput.click();
    });

    vaultFileInput.addEventListener("change", async () => {
        let currentFile = 0;
        let totalFiles = vaultFileInput.files.length;

        const startUpload = async idx => {
            await uploadFile(vaultFileInput.files[idx], idx, totalFiles, async success => {
                if (success && idx < totalFiles - 1) {
                    await startUpload(idx + 1);
                }
            });
        }

        await startUpload(currentFile);
    })

    vaultFileInput.addEventListener("click touchstart", () => {
        vaultFileInput.value = "";
    });

    let newFolderBtn = document.getElementById("new-vault-folder") as HTMLButtonElement;
    folderDialog = document.getElementById("folder-dialog");
    newFolderBtn.addEventListener("click", () => {
        let name = document.getElementById("folder-name") as HTMLInputElement;
        name.value = "";
        folderDialog.showModal();
    });

    setupFolderDialog();
    showStorageBar("", 0);

    document.addEventListener("click", event => {
        let target = (event.target as HTMLElement)
        if (target.id.startsWith(folderIDPrefix)) {
            let pageID = target.id.split("-");
            let id = pageID[pageID.length - 1];
            loadFolder(id);
        } else if (target.id.startsWith(itemIDPrefix)) {
            let itemID = target.id.split("-");
            let id = itemID[itemID.length - 1];
            downloadFile(id);
        }
    });

    document.addEventListener("click", event => {
        if (dialogs.isDialogOpen()) {
            return;
        }

        let target = (event.target as HTMLElement)
        if (target.id.startsWith(actionIDPrefix)) {
            let itemIDParts = target.id.split("-");
            let itemID = itemIDParts[itemIDParts.length - 1];
            showActionsDialog(itemID);
        }
    });
}

/**
 * Display a dialog for the current vault password (if one was set when logging in)
 * @param yeetfileDB {YeetFileDB} - The yeetfile indexeddb instance
 */
const showVaultPassDialog = (yeetfileDB: YeetFileDB) => {
    let vaultPasswordDialog = document.getElementById(
        "vault-pass-dialog") as HTMLDialogElement;
    let cancel = document.getElementById(
        "cancel-pass") as HTMLButtonElement
    cancel.addEventListener("click", () => {
        vaultPasswordDialog.close();
        window.location.assign("/");
    });

    let submit = document.getElementById("submit-pass");
    submit.addEventListener("click", async () => {
        let passwordInput = document.getElementById(
            "vault-pass") as HTMLInputElement;
        let password = passwordInput.value;
        yeetfileDB.getVaultKeyPair(password, (privKey, pubKey) => {
            vaultPasswordDialog.close();
            privateKey = privKey;
            publicKey = pubKey;

            loadFolder(folderID);
        }, () => {
            alert("Failed to decrypt vault keys. Please check your password and try again.");
        });
    })

    vaultPasswordDialog.showModal();
}

/**
 * Update the UI to allow or disallow uploading to the current folder.
 * @param allow {boolean} - True to allow, false to disallow
 */
const allowUploads = (allow: boolean): void => {
    let uploadBtn = document.getElementById("vault-upload") as HTMLButtonElement;
    let folderBtn = document.getElementById("new-vault-folder") as HTMLButtonElement;
    uploadBtn.disabled = !allow;
    folderBtn.disabled = !allow;
}

/**
 * Grabs the folder ID segment from the current path
 * @returns {string} - The current folder ID string
 */
const getFolderID = (): string => {
    let splitPath = window.location.pathname.split("/");
    for (let i = splitPath.length - 1; i > 0; i--) {
        if (splitPath[i].length > 0 && splitPath[i] !== "vault") {
            return splitPath[i];
        }
    }

    return "";
}

/**
 * Uploads one or multiple files, indicating progress to the user.
 * @param file {File} - The file to upload
 * @param idx {number} - The index of the file being uploaded (if multiple)
 * @param total {number} - The total number of files being uploaded
 * @param callback {function(boolean)} - A callback indicating if the upload was
 * successful
 */
const uploadFile = async (
    file: File,
    idx: number,
    total: number,
    callback: (success: boolean) => void,
) => {
    showFileIndicator("");

    if (total > 1) {
        setVaultMessage(`Uploading ${file.name}... (${idx + 1} / ${total})`);
    } else {
        setVaultMessage(`Uploading ${file.name}...`);
    }

    pauseInteractions = true;

    let key = await crypto.generateRandomKey();
    let protectedKey = await encryptData(key);
    let importedKey = await crypto.importKey(key);

    let encryptedName = await crypto.encryptString(importedKey, file.name);
    let hexName = toHexString(encryptedName);
    let metadata = new interfaces.VaultUpload({
        name: hexName,
        length: file.size,
        chunks: getNumChunks(file.size),
        folderID: folderID,
        protectedKey: Array.from(protectedKey),
    });

    transfer.uploadVaultMetadata(metadata, id => {
        transfer.uploadVaultChunks(id, file, importedKey, finished => {
            pauseInteractions = !finished;
            if (finished) {
                if (idx + 1 === total) {
                    let uploadName = "files";
                    if (total === 1) {
                        uploadName = file.name;
                    }
                    showStorageBar(`Finished uploading ${uploadName}!`, file.size);
                } else {
                    showStorageBar("", file.size);
                }
                loadFolder(folderID);
                callback(true);
            }
        }, errorMessage => {
            pauseInteractions = false;
            callback(false);
            alert(errorMessage);
            showStorageBar("", 0);
        });
    }, () => {
        pauseInteractions = false;
        callback(false);
        showStorageBar("", 0);
    });
}

/**
 * Decrypt encrypted file/folder data using either RSA (root folder) or AES
 * (any subfolder)
 * @param data {Uint8Array} - The data to decrypt
 * @returns {Promise<Uint8Array>} - The decrypted chunk of data
 */
const decryptData = async (data: Uint8Array): Promise<Uint8Array> => {
    if (!folderKey || folderKey.length === 0) {
        console.log("RSA decrypt")
        return await crypto.decryptRSA(privateKey, data);
    } else {
        console.log("AES decrypt")
        return await crypto.decryptChunk(folderKey, data);
    }
}

/**
 * Encrypt file/folder data using either RSA (root folder only) or AES (any
 * subfolder)
 * @param data {Uint8Array} - The data to encrypt
 */
const encryptData = async (data: Uint8Array): Promise<Uint8Array> => {
    if (!folderKey || folderKey.length === 0) {
        return await crypto.encryptRSA(publicKey, data);
    } else {
        return await crypto.encryptChunk(folderKey, data);
    }
}

/**
 * Download a vault file by file ID
 * @param id {string} - The ID of the file to download
 */
const downloadFile = (id: string): void => {
    if (pauseInteractions) {
        return;
    }

    pauseInteractions = true;
    showFileIndicator("");
    setVaultMessage("Downloading...");

    let xhr = new XMLHttpRequest();
    let url = Endpoints.format(Endpoints.DownloadVaultFileMetadata, id);
    xhr.open("GET", url, true);
    xhr.setRequestHeader('Content-Type', 'application/json');

    xhr.onreadystatechange = async () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            let download = new interfaces.VaultDownloadResponse(xhr.responseText);
            let itemKey = await decryptData(download.protectedKey);
            let tmpKey = await crypto.importKey(itemKey);
            let name = await crypto.decryptString(tmpKey, hexToBytes(download.name));

            setVaultMessage(`Downloading ${name}...`);

            transfer.downloadVaultFile(name, download, tmpKey, finished => {
                pauseInteractions = !finished;
                if (finished) {
                    showStorageBar("", 0);
                }
            }, () => {
                alert("Error downloading file!");
            });
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            alert(`Error ${xhr.status}: ${xhr.responseText}`);
            showStorageBar("", 0);
        }
    };

    xhr.send();
}

/**
 * Fetches the contents of a folder using its folder ID and displays the contents
 * of the folder in the vault view table.
 * @param newFolderID {string} - The new folder ID to fetch
 */
const loadFolder = (newFolderID: string) => {
    if (pauseInteractions) {
        return;
    }

    vaultFolders = {};
    vaultItems = {};
    folderID = newFolderID;
    folderKey = null;

    let tableBody = document.getElementById("table-body");
    tableBody.innerHTML = "";
    window.history.pushState(folderID, "", `/vault/${folderID}`);

    fetchVault(folderID, async (data: interfaces.VaultFolderResponse) => {
        subfolderParentID = data.folder.refID;
        allowUploads(data.folder.canModify);
        if (!data.keySequence || data.keySequence.length === 0) {
            // In root level vault (everything is decrypted with the user's
            // private key, since content shared with them is encrypted with
            // their public key and ends up in their root folder).
            await loadVault(data);
        } else {
            // In sub folder, need to iterate through key sequence
            folderKey = await unwindKeys(data.keySequence);
            await loadVault(data);
        }
    });
}

const unwindKeys = async (keySequence: Uint8Array[]) => {
    let parentKey;
    for (let i = 0; i < keySequence.length; i++) {
        console.log(keySequence[i]);
        if (!parentKey) {
            let protectedKey = keySequence[i];
            parentKey = await crypto.decryptRSA(privateKey, protectedKey);
            continue;
        }

        let parentKeyImport = await crypto.importKey(parentKey);
        let protectedKey = keySequence[i];
        parentKey = await crypto.decryptChunk(parentKeyImport, protectedKey);
    }

    return await crypto.importKey(parentKey);
}

const fetchVault = (
    folderID: string,
    callback: (response: interfaces.VaultFolderResponse) => void,
) => {
    let endpoint = Endpoints.format(Endpoints.VaultFolder, folderID);
    fetch(endpoint)
        .then((response) => response.text())
        .then((data) => {
            callback(new interfaces.VaultFolderResponse(data));
        })
        .catch((error) => {
            console.error("Error fetching vault: ", error);
        });
}

const loadVault = async (data) => {
    vaultItems = {};
    vaultFolders = {};

    let vaultStatus = document.getElementById("vault-status");
    if (data.folder.name.length > 0) {
        crypto.decryptString(folderKey, hexToBytes(data.folder.name)).then(folderName => {
            vaultStatus.innerHTML = `${vaultHome} | <a id="${folderIDPrefix}-${data.folder.parentID}" href="#">← Back</a> / ${folderName}`
        }).catch(() => {
            vaultStatus.innerHTML = "[decryption error]";
        });
    } else {
        vaultStatus.innerHTML = vaultHome;
    }

    let tableBody = document.getElementById("table-body");
    let folders = data.folders;
    let items = data.items;

    for (let i = 0; i < folders.length; i++) {
        let folder = folders[i];
        let subFolderKey = await decryptData(folder.protectedKey);
        let tmpKey = await crypto.importKey(subFolderKey);
        folder.name = await crypto.decryptString(tmpKey, hexToBytes(folder.name));

        vaultFolders[folder.refID] = {
            id: folder.id,
            name: folder.name,
            encKey: folder.protectedKey,
            key: tmpKey,
            tag: folder.linkTag,
            owned: folder.isOwner,
            canModify: folder.canModify,
            refID: folder.refID,
        };

        let row = await generateFolderRow(folder);
        tableBody.innerHTML += row;
    }

    for (let i = 0; i < items.length; i++) {
        let item = items[i];
        let itemKey = await decryptData(item.protectedKey);
        let tmpKey = await crypto.importKey(itemKey);
        item.name = await crypto.decryptString(tmpKey, hexToBytes(item.name));

        vaultItems[item.refID] = {
            id: item.id,
            name: item.name,
            encKey: item.protectedKey,
            key: tmpKey,
            owned: item.isOwner,
            canModify: item.canModify,
            refID: item.refID,
        };

        let row = await generateItemRow(item);
        tableBody.innerHTML += row;
    }

    for (let i = 0; i < gapFill - (folders.length + items.length); i++) {
        tableBody.innerHTML += emptyRow;
    }
}

const addTableRow = row => {
    let tableBody = document.getElementById("table-body");
    tableBody.innerHTML = row + tableBody.innerHTML;
}

const setupFolderDialog = () => {
    let cancelUploadBtn = document.getElementById("cancel-folder");
    cancelUploadBtn.addEventListener("click", () => dialogs.closeDialog(folderDialog));

    let submitUploadBtn = document.getElementById("submit-folder");
    submitUploadBtn.addEventListener("click", () => {
        let nameInput = document.getElementById("folder-name") as HTMLInputElement;
        let folderName = nameInput.value;
        if (folderName.length === 0) {
            alert("Folder name must be > 0 characters");
            return;
        }
        createNewFolder(folderName).then(() => dialogs.closeDialog(folderDialog));
    });
}

const generateFolderRow = async (item) => {
    let classes = item.sharedBy.length > 0 ? "shared-link" : "folder-link";
    let link = `<a id="${folderIDPrefix}-${item.refID}" class="${classes}" href="#">${item.name}/</a>`
    return generateRow(
        link,
        item.name,
        "",
        formatDate(item.modified),
        item.refID,
        true,
        item.sharedWith,
        item.sharedBy);
}

const generateItemRow = async (item) => {
    let classes = item.sharedBy.length > 0 ? "shared-link" : "file-link";
    let id = `${itemIDPrefix}-${item.refID}`;
    let link = `<a data-testid="${id}" id="${id}" class="${classes}" href="#">${item.name}</a>`
    return generateRow(
        link,
        item.name,
        calcFileSize(item.size - constants.TotalOverhead),
        formatDate(item.modified),
        item.refID,
        false,
        item.sharedWith,
        item.sharedBy);
}

const generateRow = (link, name, size, modified, id, isFolder, sharedWith, sharedBy) => {
    let iconClasses = sharedBy ? "small-icon shared-icon" : "small-icon";
    let icon = `<img class="${iconClasses}" src="/static/icons/file.svg">`
    if (isFolder) {
        icon = `<img class="${iconClasses} accent-icon" src="/static/icons/folder.svg">`
    }

    let sharedIcon = generateSharedWithIcon(id, sharedWith);
    // let linkedIcon = isLinked ? `<img id="${id}-linked" class="small-icon" src="/static/icons/link.svg">` : ""
    let sharedByIndicator = sharedBy ? `<br><img class="small-icon shared-icon" src="/static/icons/owner.svg">&nbsp;${sharedBy}` : ""

    let idStr = `${actionIDPrefix}-${id}`
    return `<tr id="${id}-row">
        <td>${icon} ${link} ${sharedIcon} ${sharedByIndicator}</td>
        <td>${size}</td>
        <td id="${id}-modified">${modified}</td>
        <td class="action-icon" data-testid=${idStr} id="${idStr}">⋮</td>
    </tr>`
}

const generateSharedWithIcon = (id, sharedWithCount) => {
    let div = `<div class="shared-with-div" id="${id}-${sharedWithSuffix}">`
    if (sharedWithCount === 0) {
        return div + "</div>";
    }

    return `${div}<img class="small-icon" src="/static/icons/share.svg"> (${sharedWithCount})</div>`
}

const updateRow = (id, isFolder, name) => {
    let prefix = isFolder ? folderIDPrefix : itemIDPrefix;
    let nameID = `${prefix}-${id}`;
    let modifiedID = `${id}-modified`;
    document.getElementById(nameID).innerText = name + (isFolder ? "/" : "");
    document.getElementById(modifiedID).innerText = formatDate(Date());
}

const removeRow = (id) => {
    let rowID = `${id}-row`;
    document.getElementById(rowID).remove();
}

const showActionsDialog = (id) => {
    let isFolder = false;
    let actionsDialog = document.getElementById("actions-dialog") as HTMLDialogElement;
    let title = document.getElementById("actions-dialog-title") as HTMLHeadingElement;
    let item;

    if (vaultFolders[id]) {
        item = vaultFolders[id];
        title.innerText = "Folder: " + item.name;
        isFolder = true;
    } else if (vaultItems[id]) {
        item = vaultItems[id];
        title.innerText = "File: " + item.name;
    }

    let actionDownload = document.getElementById("action-download");
    actionDownload.style.display = isFolder ? "none" : "flex";
    actionDownload.addEventListener("click", event => {
        event.stopPropagation();
        dialogs.closeDialog(actionsDialog);
        downloadFile(id);
    });

    // TOmaybeDO
    let actionSend = document.getElementById("action-send");
    //actionSend.style.display = isFolder ? "none" : "flex";
    actionSend.style.display = "none";
    actionSend.addEventListener("click", event => { });

    let actionRename = document.getElementById("action-rename");
    if (item.canModify) {
        actionRename.style.display = "flex";
        actionRename.addEventListener("click", (event) => {
            event.stopPropagation();
            showRenameDialog(id, isFolder);
            dialogs.closeDialog(actionsDialog);
        });
    } else {
        actionRename.style.display = "none";
    }

    let actionLink = document.getElementById("action-link");
    actionLink.style.display = "none";
    // if (item.owned) {
    //     actionLink.style.display = "flex";
    //     actionLink.addEventListener("click", event => {
    //         event.stopPropagation();
    //         showLinkDialog(id, isFolder);
    //         dialogs.closeDialog(actionsDialog);
    //     });
    // }

    let actionShare = document.getElementById("action-share");
    if (item.owned) {
        actionShare.style.display = "flex";
        actionShare.addEventListener("click", async event => {
            event.stopPropagation();
            let itemKey = isFolder ? vaultFolders[id].encKey : vaultItems[id].encKey;
            let itemKeyRaw = await decryptData(itemKey);
            await dialogs.showShareDialog(id, itemKeyRaw, isFolder, signal => {
                if (signal !== dialogs.DialogSignal.Cancel) {
                    transfer.getSharedUsers(id, isFolder).then(response => {
                        let icon = generateSharedWithIcon(id,
                            response ?
                                (response as Array<JSON>).length :
                                0);
                        document.getElementById(`${id}-${sharedWithSuffix}`).innerHTML = icon;
                    })
                }
            });
            dialogs.closeDialog(actionsDialog);
        });
    } else {
        actionShare.style.display = "none";
    }

    let actionDelete = document.getElementById("action-delete");
    if (item.owned) {
        actionDelete.style.display = "flex";
        actionDelete.addEventListener("click", event => {
            event.stopPropagation();

            if (!isFolder && confirm("Are you sure you want to delete this file?")) {
                dialogs.closeDialogs();
                deleteVaultContent(id, item.name, isFolder, item.refID, async response => {
                    removeRow(id);
                    let responseJSON = await response.json();
                    showStorageBar("", responseJSON.freedSpace * -1);
                });
            } else if (isFolder && confirm("Are you sure you want to delete this folder? " +
                "This will delete all files in the folder permanently.")) {
                dialogs.closeDialogs();
                deleteVaultContent(id, item.name, isFolder, "", () => {
                    location.reload();
                });
            }
        });
    } else {
        actionDelete.style.display = "none";
    }

    let actionRemove = document.getElementById("action-remove");
    if (folderID.length === 0 && !item.owned) {
        actionRemove.addEventListener("click", event => {
            event.stopPropagation();
            if (confirm("Are you sure you want to remove this item? " +
                "The owner will need to re-share this with you if you need access again.")) {
                dialogs.closeDialogs();
                deleteVaultContent(id, item.name, isFolder, item.id, () => {
                    location.reload();
                });
            }
        });
    } else {
        actionRemove.style.display = "none";
    }

    let cancel = document.getElementById("cancel-action");
    cancel.addEventListener("click", () => dialogs.closeDialog(actionsDialog));

    actionsDialog.showModal();
}

const showRenameDialog = (id, isFolder) => {
    let renameDialog = document.getElementById("rename-dialog") as HTMLDialogElement;
    let dialogTitle = document.getElementById("rename-title") as HTMLHeadingElement;
    dialogTitle.innerText = isFolder ? "Rename Folder" : "Rename File";

    let newNameInput = document.getElementById("new-name") as HTMLInputElement;
    newNameInput.value = isFolder ? vaultFolders[id].name : vaultItems[id].name;

    let cancelBtn = document.getElementById("cancel-rename");
    cancelBtn.addEventListener("click", () => {
        dialogs.closeDialog(renameDialog);
        showActionsDialog(id);
    });

    let submitBtn = document.getElementById("submit-rename") as HTMLButtonElement;
    submitBtn.addEventListener("click", async () => {
        submitBtn.disabled = true;
        await renameItem(id, isFolder, newNameInput.value);
    });

    renameDialog.showModal();
}

const renameItem = async (id, isFolder, newName) => {
    let key;
    if (isFolder) {
        key = vaultFolders[id].key;
    } else {
        key = vaultItems[id].key;
    }

    let newNameEncrypted = await crypto.encryptString(key, newName);
    let endpoint = isFolder ?
        Endpoints.format(Endpoints.VaultFolder, id) :
        Endpoints.format(Endpoints.VaultFile, id);
    fetch(endpoint, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: toHexString(newNameEncrypted) })
    }).then(response => {
        if (response.ok) {
            dialogs.closeDialogs();
            if (isFolder) {
                vaultFolders[id].name = newName;
                updateRow(id, isFolder, newName);
            } else {
                vaultItems[id].name = newName;
                updateRow(id, isFolder, newName);
            }

            showActionsDialog(id);
        } else {
            alert("Error renaming file");
        }
    }).catch(error => {
        alert("Error renaming file: " + error);
    });
}

const deleteVaultContent = (id, name, isFolder, sharedID, callback) => {
    let modID = sharedID !== id ? sharedID : id;
    let endpoint = isFolder ?
        Endpoints.format(Endpoints.VaultFolder, modID) :
        Endpoints.format(Endpoints.VaultFile, modID);

    pauseInteractions = true;
    showFileIndicator(`Deleting ${name}...`);

    let sharedParam = sharedID !== id ? "?shared=true" : "";

    fetch(`${endpoint}${sharedParam}`, {
        method: "DELETE"
    }).then(response => {
        pauseInteractions = false;
        if (response.ok) {
            callback(response);
        } else {
            alert("Error deleting item");
            showStorageBar("", 0);
        }
    }).catch(error => {
        alert("Error deleting item: " + error);
    });
}

const createNewFolder = async (folderName) => {
    let xhr = new XMLHttpRequest();
    xhr.open("POST", Endpoints.format(Endpoints.VaultFolder, ""), false);
    xhr.setRequestHeader("Content-Type", "application/json");

    let newFolderKey = await crypto.generateRandomKey();
    let newFolderKeyImported = await crypto.importKey(newFolderKey);
    let encFolderKey = await encryptData(newFolderKey);
    let encName = await crypto.encryptString(newFolderKeyImported, folderName);
    let encNameEncoded = toHexString(encName);

    xhr.onreadystatechange = async () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            let response = new interfaces.NewFolderResponse(xhr.responseText);
            vaultFolders[response.id] = {
                name: folderName,
                encKey: encFolderKey,
                key: newFolderKey,
                tag: "",
                owned: true,
                canModify: true,
                refID: response.id,
            };

            let row = await generateFolderRow({
                id: response.id,
                refID: response.id,
                name: folderName,
                modified: new Date().toLocaleString(),
                sharedWith: 0,
                sharedBy: "",
                linkTag: "",
            });
            addTableRow(row);
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            alert(`Error ${xhr.status}: ${xhr.responseText}`);
        }
    };

    xhr.send(JSON.stringify({
        name: encNameEncoded,
        parentID: folderID,
        protectedKey: Array.from(encFolderKey),
    }));
}

/**
 * Updates the storage indicator on the vault page with a change in storage
 * amount, or a temporary message.
 * @param tmpMessage {string} - The temporary message to display to the user
 * @param usedStorageDiff {number} - The difference in used storage (can be negative)
 * @returns {void}
 */
const showStorageBar = (tmpMessage: string, usedStorageDiff: number): void => {
    let storageBar = document.getElementById("storage-bar") as HTMLProgressElement;
    let itemBar = document.getElementById("item-bar") as HTMLProgressElement;
    let vaultMessage = document.getElementById("vault-message") as HTMLSpanElement;

    storageBar.style.display = "inherit";
    itemBar.style.display = "none";

    if (storageBar.max <= 1 && storageBar.value === 0) {
        storageBar.style.display = "none";
        vaultMessage.innerHTML = "<span>" +
            "<a href='/account'>Membership</a> required for vault storage" +
            "</span>";
        return;
    }

    if (usedStorageDiff) {
        storageBar.value += usedStorageDiff;
    }

    let available = calcFileSize(storageBar.max);
    let used = calcFileSize(storageBar.value);

    if (tmpMessage) {
        vaultMessage.innerText = tmpMessage;
        setTimeout(() => {
            vaultMessage.innerText = `Storage: ${used} / ${available}`
        }, 3000);
    } else {
        vaultMessage.innerText = `Storage: ${used} / ${available}`
    }
}

/** Hides the storage bar and displays a file upload indicator in its place.
 * @param msg {string} - A message to display alongside the indicator
 * @returns {void}
 */
const showFileIndicator = (msg: string) => {
    let storageBar = document.getElementById("storage-bar");
    let itemBar = document.getElementById("item-bar");

    storageBar.style.display = "none";
    itemBar.style.display = "inherit";

    if (msg) {
        let vaultMessage = document.getElementById("vault-message");
        vaultMessage.innerText = msg;
    }
}

const setVaultMessage = msg => {
    let vaultMessage = document.getElementById("vault-message");
    vaultMessage.innerHTML = `<img class="small-icon progress-spinner" src="/static/icons/progress.svg">${msg}`;
}

if (document.readyState !== 'loading') {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}
