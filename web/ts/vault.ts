import * as crypto from "./crypto.js";
import * as dialogs from "./dialogs.js";
import * as transfer from "./transfer.js";
import * as constants from "./constants.js";
import { Endpoints } from "./endpoints.js";
import * as interfaces from "./interfaces.js"
import { YeetFileDB } from "./db.js";
import * as render from "./render.js";
import * as fragments from "./fragments.js";
import {DeleteResponse, VaultItem} from "./interfaces.js";

const gapFill = 9;
const closeFileID = "close-file";
const actionIDPrefix = "action";
const folderIDPrefix = "load-folder";
const itemIDPrefix = "load-item";
const sharedWithSuffix = "sharedwith";
const folderRowSuffix = "folder-row";
const fileRowSuffix = "file-row";

const emptyRow = `<tr class="blank-row"><td colspan="4"></td></tr>`
const vaultHome = `<a id="${folderIDPrefix}-" href="#">Home</a>`;
const closeFile = `<a id="${closeFileID}" href="#">Close</a>`;
const folderPlaceholder = `<a href="/vault">← Back</a> / ...`

class VaultViewFile {
    decName: string;
    key: CryptoKey;
    [key: string]: any;

    constructor(item: interfaces.VaultItem, key: CryptoKey, decName: string) {
        Object.assign(this, item);
        this.key = key;
        this.decName = decName;
    }
}

class VaultViewFolder {
    decName: string;
    key: CryptoKey;
    [key: string]: any;

    constructor(item: interfaces.VaultFolder, key: CryptoKey, decName: string) {
        Object.assign(this, item);
        this.key = key;
        this.decName = decName;
    }
}

type FolderCache = {
    [key: string]: interfaces.VaultFolderResponse;
};

type CurrentFiles = {
    [key: string]: VaultViewFile;
}

type CurrentFolders = {
    [key: string]: VaultViewFolder;
}

enum View {
    File = 0,
    Folder
}

let cache: FolderCache = {};
let folderStatus: string = "";

let folderDialog: HTMLDialogElement;
let subfolderParentID: string;
let privateKey: CryptoKey;
let publicKey: CryptoKey;
let folderID = "";
let folderKey;

let pauseInteractions = false;

let currentFiles: CurrentFiles = {};
let currentFolders: CurrentFolders = {};

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
            yeetfileDB.getVaultKeyPair("", false, (privKey: CryptoKey, pubKey: CryptoKey) => {
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

    vaultFileInput.addEventListener("change", uploadUserSelectedFiles);
    vaultFileInput.addEventListener("click touchstart", () => {
        vaultFileInput.value = "";
    });

    let newFolderBtn = document.getElementById("new-vault-folder") as HTMLButtonElement;
    folderDialog = document.getElementById("folder-dialog") as HTMLDialogElement;
    newFolderBtn.addEventListener("click", () => {
        let name = document.getElementById("folder-name") as HTMLInputElement;
        name.value = "";
        folderDialog.showModal();
    });

    setupFolderDialog();
    showStorageBar("", 0);
    document.addEventListener("click", clickListener, { passive: true });
}

/**
 * Handles events when a user clicks on an element in the vault view (file,
 * folder, actions button, etc).
 * @param event {MouseEvent}
 */
const clickListener = (event: MouseEvent) => {
    if (dialogs.isDialogOpen() || (event.target as Element).closest("code")) {
        return;
    }

    let target = (event.target as HTMLElement);
    if (target.id.startsWith(actionIDPrefix)) {
        let itemIDParts = target.id.split("-");
        let itemID = itemIDParts[itemIDParts.length - 1];
        showActionsDialog(itemID);
    } else if (target.id.startsWith(folderIDPrefix)) {
        let pageID = target.id.split("-");
        let id = pageID[pageID.length - 1];
        loadFolder(id);
    } else if (target.id.startsWith(itemIDPrefix)) {
        let itemID = target.id.split("-");
        let id = itemID[itemID.length - 1];
        loadFile(id);
    } else if (target.id === closeFileID) {
        closeFileView();
    }
}

/**
 * Initiates file upload after files are selected by the user
 */
const uploadUserSelectedFiles = async () => {
    let vaultFileInput = document.getElementById("file-input") as HTMLInputElement;
    let currentFile = 0;
    let totalFiles = vaultFileInput.files.length;

    const startUpload = async idx => {
        await uploadFile(
            vaultFileInput.files[idx],
            idx,
            totalFiles,
            async (success, file, view ) => {
                if (success) {
                    let row = await generateFileRow(view);
                    currentFiles[file.id] = view;
                    cache[folderID].items.unshift(file);
                    insertFileRow(row);
                    if (idx < totalFiles - 1) {
                        await startUpload(idx + 1);
                    }
                }
            });
    }

    await startUpload(currentFile);
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
        yeetfileDB.getVaultKeyPair(password, false, (privKey: CryptoKey, pubKey: CryptoKey) => {
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
 * @param callback {function(boolean, VaultItem, VaultViewFile)} - A callback
 * indicating if the upload was successful, and if so, the file and view class
 * for that file
 */
const uploadFile = async (
    file: File,
    idx: number,
    total: number,
    callback: (success: boolean, item: VaultItem, file: VaultViewFile) => void,
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

                let item = new VaultItem();
                item.id = id;
                item.refID = id;
                item.name = hexName;
                item.size = file.size + constants.TotalOverhead;
                item.modified = new Date();
                item.protectedKey = protectedKey;
                item.sharedBy = "";
                item.sharedWith = 0;
                item.canModify = cache[folderID].folder.canModify;
                item.isOwner = cache[folderID].folder.isOwner;

                let view = new VaultViewFile(item, importedKey, file.name);
                callback(true, item, view);
            }
        }, errorMessage => {
            pauseInteractions = false;
            callback(false, undefined, undefined);
            alert(errorMessage);
            showStorageBar("", 0);
        });
    }, () => {
        pauseInteractions = false;
        callback(false, undefined, undefined);
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
        return await crypto.decryptRSA(privateKey, data);
    } else {
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

const setTableLoading = (loading: boolean) => {
    let loadingHeader = document.getElementById("loading-header");
    let tableHeader = document.getElementById("table-header");

    if (loading) {
        loadingHeader.className = "visible";
        tableHeader.className = "hidden";
    } else {
        loadingHeader.className = "hidden";
        tableHeader.className = "visible";
    }
}

const setView = (view: View) => {
    let fileDiv = document.getElementById("vault-file-div");
    let folderDiv = document.getElementById("vault-items-div");

    if (view === View.File) {
        fileDiv.className = "visible";
        folderDiv.className = "hidden";
    } else if (view === View.Folder) {
        fileDiv.innerHTML = fragments.VaultFileViewDiv();
        fileDiv.className = "hidden";
        folderDiv.className = "visible";
    }
}

/**
 * Update the status bar with new HTML
 * @param contents {string}
 */
const setStatus = (contents: string) => {
    let vaultStatus = document.getElementById("vault-status");
    vaultStatus.innerHTML = contents;
}

/**
 * Resets the decrypted file and folder cache
 */
const emptyCurrentItems = () => {
    currentFiles = {};
    currentFolders = {};
}

/**
 * Fetches the contents of a folder using its folder ID and displays the contents
 * of the folder in the vault view table.
 * @param newFolderID {string} - The new folder ID to fetch
 */
const loadFolder = async (newFolderID: string) => {
    if (pauseInteractions) {
        return;
    }

    setTableLoading(true);
    setView(View.Folder);

    let tableBody = document.getElementById("table-body");
    tableBody.innerHTML = "";

    folderID = newFolderID;

    let folderPath = Endpoints.format(Endpoints.HTMLVaultFolder, folderID);
    window.history.pushState(folderID, "", folderPath);

    const loadFolderData = async (data: interfaces.VaultFolderResponse) => {
        emptyCurrentItems();
        folderKey = null;

        subfolderParentID = data.folder.refID;
        allowUploads(data.folder.canModify);
        if (!data.keySequence || data.keySequence.length === 0) {
            // In root level vault (everything is decrypted with the user's
            // private key, since content shared with them is encrypted with
            // their public key and ends up in their root folder).
            await loadVault(data);
        } else {
            // In sub folder, need to iterate through key sequence
            folderKey = await crypto.unwindKeys(privateKey, data.keySequence);
            await loadVault(data);
        }
    }

    if (cache[folderID]) {
        await loadFolderData(cache[folderID]);
    } else {
        fetchVault(folderID, async (data: interfaces.VaultFolderResponse) => {
            cache[folderID] = data;
            await loadFolderData(data);
        });
    }
}

/**
 * Displays file metadata and displays a preview of the file.
 * @param fileID
 */
const loadFile = (fileID: string) => {
    if (pauseInteractions) {
        return;
    }

    setStatus(closeFile);
    setView(View.File);

    let file = currentFiles[fileID];
    if (!file) {
        alert("Unable to open file!");
        loadFolder(folderID);
        return;
    }

    let headerDiv = document.getElementById("vault-file-header");
    let htmlDiv = document.getElementById("vault-file-content");
    let textOnly = document.getElementById("vault-text-content");
    let textWrapper = document.getElementById("vault-text-wrapper");

    headerDiv.innerHTML = "";
    htmlDiv.innerHTML = "";
    textOnly.innerText = "";

    generateItemView(file, header => {
        headerDiv.innerHTML = header;
    }, html => {
        htmlDiv.innerHTML = html;
        textWrapper.className = "hidden";
        htmlDiv.className = "visible";
    }, text => {
        textOnly.innerText = text;
        htmlDiv.className = "hidden";
        textWrapper.className = "visible";
    });
}

/**
 * Closes the file viewer, re-opens the folder view, and resets the status
 * bar to the current folder status.
 */
const closeFileView = () => {
    setView(View.Folder);
    setStatus(folderStatus);
}

/**
 * Fetches the current vault folder contents by vault folder ID. Fires a
 * callback containing the vault folder response when finished.
 * @param folderID {string}
 * @param callback {interfaces.VaultFolderResponse}
 */
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

/**
 * Uses a VaultFolderResponse to load the contents into the view.
 * @param data {interfaces.VaultFolderResponse}
 */
const loadVault = async (data: interfaces.VaultFolderResponse) => {
    emptyCurrentItems();

    if (data.folder.name.length > 0) {
        crypto.decryptString(folderKey, hexToBytes(data.folder.name)).then(folderName => {
            let folderLink = `${vaultHome} | <a id="${folderIDPrefix}-${data.folder.parentID}" href="#">← Back</a> / ${folderName}`;
            folderStatus = folderLink;
            setStatus(folderLink);
        }).catch(() => {
            setStatus("[decryption error]");
        });
    } else {
        setStatus(vaultHome);
        folderStatus = vaultHome;
    }

    let tableBody = document.getElementById("table-body");
    let folders = data.folders;
    let items = data.items;

    for (let i = 0; i < folders.length; i++) {
        let folder = folders[i];
        let subFolderKey = await decryptData(folder.protectedKey);
        let tmpKey = await crypto.importKey(subFolderKey);
        let decName = await crypto.decryptString(tmpKey, hexToBytes(folder.name));

        let vaultFolder = new VaultViewFolder(folder, tmpKey, decName);
        currentFolders[folder.refID] = vaultFolder;
        let row = await generateFolderRow(vaultFolder);
        tableBody.innerHTML += row;
    }

    for (let i = 0; i < items.length; i++) {
        let item = items[i];
        let itemKey = await decryptData(item.protectedKey);
        let tmpKey = await crypto.importKey(itemKey);
        let decName = await crypto.decryptString(tmpKey, hexToBytes(item.name));

        let vaultFile = new VaultViewFile(item, tmpKey, decName);
        currentFiles[item.refID] = vaultFile;
        let row = await generateFileRow(vaultFile);
        tableBody.innerHTML += row;
    }

    for (let i = 0; i < gapFill - (folders.length + items.length); i++) {
        tableBody.innerHTML += emptyRow;
    }

    setTableLoading(false);
}

/**
 * Prepends an HTML string `tr` element to the vault table body. Note that this
 * should only be used for folders, since new ones should always be at the top.
 * @param row {string}
 */
const addTableRow = (row: string) => {
    let tableBody = document.getElementById("table-body");
    tableBody.innerHTML = row + tableBody.innerHTML;
}

/**
 * Inserts an HTML string `tr` element into the vault table element below all
 * folders but above any existing files.
 * @param row {string}
 */
const insertFileRow = (row: string) => {
    let tableBody = document.getElementById("table-body");
    let rows = tableBody.getElementsByTagName("tr");

    let tempContainer = document.createElement("tbody");
    tempContainer.innerHTML = row;
    let rowNode = tempContainer.firstElementChild;

    if (rows.length === 0) {
        tableBody.innerHTML = row;
        return;
    }

    for (let i = 0; i < rows.length; i++) {
        if (!rows[i].id.endsWith(fileRowSuffix) && rows[i].className !== "blank-row") {
            continue;
        }

        tableBody.insertBefore(rowNode, rows[i]);
        return;
    }

    tableBody.innerHTML += row;
}

/**
 * Sets up event listeners for the new folder dialog.
 */
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

/**
 * Creates an HTML `tr` element string using properties of the provided
 * VaultViewFolder element.
 * @param folder {VaultViewFolder}
 */
const generateFolderRow = async (folder: VaultViewFolder) => {
    let classes = folder.sharedBy.length > 0 ? "shared-link" : "folder-link";
    let link = `<a id="${folderIDPrefix}-${folder.refID}" class="${classes}" href="#">${folder.decName}/</a>`
    return generateRow(
        link,
        folder.decName,
        "",
        formatDate(folder.modified),
        folder.refID,
        true,
        folder.sharedWith,
        folder.sharedBy);
}

/**
 * Creates an HTML `tr` element string using properties of the provided
 * VaultViewFile element.
 * @param file {VaultViewFile}
 */
const generateFileRow = async (file: VaultViewFile) => {
    let classes = file.sharedBy.length > 0 ? "shared-link" : "file-link";
    let id = `${itemIDPrefix}-${file.refID}`;
    let link = `<a data-testid="${id}" id="${id}" class="${classes}" href="#">${file.decName}</a>`
    return generateRow(
        link,
        file.decName,
        calcFileSize(file.size - constants.TotalOverhead),
        formatDate(file.modified),
        file.refID,
        false,
        file.sharedWith,
        file.sharedBy);
}

/**
 * Generates an HTML string `tr` element for either a file or folder, using
 * the provided values that are needed for the view.
 * @param link {string} - The `a` tag for the vault item
 * @param name {string} - The decrypted name of the vault item
 * @param size {string} - The file size string
 * @param modified {string} - The formatted date string
 * @param id {string} - The file/folder ID
 * @param isFolder {boolean} - True if folder, otherwise false
 * @param sharedWith {number} - The number of people the file/folder is shared with
 * @param sharedBy {string} - The user who shared the content with the current user
 */
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
    return `<tr id="${id}-${isFolder ? folderRowSuffix : fileRowSuffix}">
        <td>${icon} ${link} ${sharedIcon} ${sharedByIndicator}</td>
        <td>${size}</td>
        <td id="${id}-modified">${modified}</td>
        <td class="action-icon" data-testid=${idStr} id="${idStr}">⋮</td>
    </tr>`
}

/**
 * Generates an icon indicating the number of users a file/folder has been
 * shared with.
 * @param id {string}
 * @param sharedWithCount {number}
 */
const generateSharedWithIcon = (id: string, sharedWithCount: number) => {
    let div = `<div class="shared-with-div" id="${id}-${sharedWithSuffix}">`
    if (sharedWithCount === 0) {
        return div + "</div>";
    }

    return `${div}<img class="small-icon" src="/static/icons/share.svg"> (${sharedWithCount})</div>`
}

/**
 * Generates a view of the selected VaultViewFile, displaying the contents in
 * a temporary div. Provides three callbacks for the caller:
 * - header: The header content to display above the file
 * - html: The HTML representation of the file (if applicable)
 * - text: The plaintext representation of the file (if applicable)
 *
 * The `html` callback is used if the file is an image, audio file, pdf, etc.
 * Otherwise the `text` callback is used.
 * @param file {VaultViewFile}
 * @param header {callback(string)}
 * @param html {callback(string)}
 * @param text {callback(string)}
 */
const generateItemView = (
    file: VaultViewFile,
    header: (string) => void,
    html: (string) => void,
    text: (string) => void,
) => {
    let name = file.decName;
    let date = formatDate(file.modified);
    let size = calcFileSize(file.size - constants.TotalOverhead);
    let htmlHead = `<p class="no-top-margin">${name} | ${size} | ${date}</p><hr>`;
    header(htmlHead+fragments.LoadingSpinner("sub"));

    let url = Endpoints.format(Endpoints.DownloadVaultFileMetadata, file.refID);
    fetch(url).then(response => {
        header(htmlHead);
        if (!response.ok) {
            text("Failed to fetch file");
            return;
        }

        response.json().then(json => {
            let metadata = new interfaces.VaultDownloadResponse(json);
            if (metadata.chunks > 5) {
                text("File too large to preview");
                return;
            }

            let bytes;
            const fetchChunk = (chunkNum: number) => {
                let endpoint = Endpoints.DownloadVaultFileData;
                let chunkURL = Endpoints.format(endpoint, metadata.id, `${chunkNum}`);
                transfer.fetchSingleChunk(chunkURL, file.key, chunk => {
                    if (!bytes) {
                        bytes = chunk;
                    } else {
                        let combinedBuffer = new ArrayBuffer(chunk.byteLength + bytes.byteLength);
                        let combinedArray = new Uint8Array(combinedBuffer);

                        combinedArray.set(new Uint8Array(bytes), 0);
                        combinedArray.set(new Uint8Array(chunk), bytes.byteLength);
                        bytes = combinedArray;
                    }

                    if (chunkNum === metadata.chunks) {
                        if (render.isNonTextFileType(file.decName)) {
                            render.renderFileHTML(file.decName, bytes, (tag, url) => {
                                html(tag);
                                let popOut = `<a target="_blank" href="${url}">Pop Out</a>`;
                                let download = `<a href="${url}" download="${name}">Download</a>`;
                                setStatus(`${closeFile} | ${popOut} | ${download}`);
                            });
                        } else {
                            let textContent = new TextDecoder().decode(bytes);
                            text(textContent);
                        }
                    } else {
                        fetchChunk(chunkNum + 1);
                    }
                }, () => {
                    text("Error downloading file content");
                });
            }

            fetchChunk(1);
        })
    });
}

/**
 * Updates the name of a file/folder
 * @param id {string} - The ID of the modified content
 * @param isFolder {boolean} - True if a folder, else false
 * @param name {string} - The new (unencrypted) name
 */
const updateRow = (id: string, isFolder: boolean, name: string) => {
    let prefix = isFolder ? folderIDPrefix : itemIDPrefix;
    let nameID = `${prefix}-${id}`;
    let modifiedID = `${id}-modified`;
    document.getElementById(nameID).innerText = name + (isFolder ? "/" : "");
    document.getElementById(modifiedID).innerText = formatDate(Date());
}

/**
 * Removes a file or folder row from the vault view table.
 * @param id {string} - The file/folder ID
 * @param isFolder {boolean} - True if a folder, else false
 */
const removeRow = (id: string, isFolder: boolean) => {
    let rowID = `${id}-${isFolder ? folderRowSuffix : fileRowSuffix}`;
    document.getElementById(rowID).remove();
}

/**
 * Removes a file or folder from the local folder cache.
 * @param id {string}
 * @param isFolder {boolean}
 */
const removeFromCache = (id: string, isFolder: boolean) => {
    if (isFolder) {
        cache[folderID].folders = cache[folderID].folders.filter(obj => {
            return obj.id !== id;
        });
    } else {
        cache[folderID].items = cache[folderID].items.filter(obj => {
            return obj.id !== id;
        });
    }
}

/**
 * Displays the actions dialog for a file or folder
 * @param id {string}
 */
const showActionsDialog = (id: string) => {
    let isFolder = false;
    let actionsDialog = document.getElementById("actions-dialog") as HTMLDialogElement;
    let title = document.getElementById("actions-dialog-title") as HTMLHeadingElement;
    let item;

    if (currentFolders[id]) {
        item = currentFolders[id];
        title.innerText = "Folder: " + item.decName;
        isFolder = true;
    } else if (currentFiles[id]) {
        item = currentFiles[id];
        title.innerText = "File: " + item.decName;
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
    // if (item.isOwner) {
    //     actionLink.style.display = "flex";
    //     actionLink.addEventListener("click", event => {
    //         event.stopPropagation();
    //         showLinkDialog(id, isFolder);
    //         dialogs.closeDialog(actionsDialog);
    //     });
    // }

    let actionShare = document.getElementById("action-share");
    if (item.isOwner) {
        actionShare.style.display = "flex";
        actionShare.addEventListener("click", async event => {
            event.stopPropagation();
            let itemKey = isFolder ? currentFolders[id].key : currentFiles[id].key;
            let itemKeyRaw = await crypto.exportKey(itemKey, "raw");
            await dialogs.showShareDialog(id, itemKeyRaw, isFolder, signal => {
                if (signal === dialogs.DialogSignal.Cancel) {
                    return;
                }
                transfer.getSharedUsers(id, isFolder).then(response => {
                    let icon = generateSharedWithIcon(id,
                        response ?
                            (response as Array<JSON>).length :
                            0);
                    document.getElementById(`${id}-${sharedWithSuffix}`).innerHTML = icon;
                });

            });
            dialogs.closeDialog(actionsDialog);
        });
    } else {
        actionShare.style.display = "none";
    }

    let actionDelete = document.getElementById("action-delete");
    if (item.isOwner) {
        actionDelete.style.display = "flex";
        actionDelete.addEventListener("click", event => {
            event.stopPropagation();

            let confirmMsg;
            if (isFolder) {
                confirmMsg = "Are you sure you want to delete this folder? " +
                    "This will delete all files in the folder permanently.";
            } else {
                confirmMsg = "Are you sure you want to delete this file?";
            }

            if (confirm(confirmMsg)) {
                dialogs.closeDialogs();
                deleteVaultContent(id, item.decName, isFolder, item.refID, response => {
                    removeRow(id, isFolder);
                    removeFromCache(id, isFolder);
                    showStorageBar("", response.freedSpace * -1);
                });
            }
        });
    } else {
        actionDelete.style.display = "none";
    }

    let actionRemove = document.getElementById("action-remove");
    if (folderID.length === 0 && !item.isOwner) {
        actionRemove.addEventListener("click", event => {
            event.stopPropagation();
            if (confirm("Are you sure you want to remove this item? " +
                "The owner will need to re-share this with you if you need access again.")) {
                dialogs.closeDialogs();
                deleteVaultContent(id, item.decName, isFolder, item.id, () => {
                    removeRow(id, isFolder);
                    removeFromCache(id, isFolder);
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

/**
 * Displays the renaming dialog for a file or folder
 * @param id {string}
 * @param isFolder {boolean}
 */
const showRenameDialog = (id: string, isFolder: boolean) => {
    let renameDialog = document.getElementById("rename-dialog") as HTMLDialogElement;
    let dialogTitle = document.getElementById("rename-title") as HTMLHeadingElement;
    dialogTitle.innerText = isFolder ? "Rename Folder" : "Rename File";

    let newNameInput = document.getElementById("new-name") as HTMLInputElement;
    newNameInput.value = isFolder ? currentFolders[id].decName : currentFiles[id].decName;

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

/**
 * Renames a file or folder to the new name
 * @param id {string}
 * @param isFolder {boolean}
 * @param newName {string}
 */
const renameItem = async (id: string, isFolder: boolean, newName: string) => {
    let key;
    if (isFolder) {
        key = currentFolders[id].key;
    } else {
        key = currentFiles[id].key;
    }

    let newNameEncrypted = await crypto.encryptString(key, newName);
    let hexName = toHexString(newNameEncrypted);
    let endpoint = isFolder ?
        Endpoints.format(Endpoints.VaultFolder, id) :
        Endpoints.format(Endpoints.VaultFile, id);
    fetch(endpoint, {
        method: "PUT",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: hexName })
    }).then(response => {
        if (response.ok) {
            dialogs.closeDialogs();
            if (isFolder) {
                currentFolders[id].name = hexName;
                currentFolders[id].decName = newName;
                updateRow(id, isFolder, newName);
            } else {
                currentFiles[id].name = hexName;
                currentFiles[id].decName = newName;
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

/**
 * Deletes a file or folder from the user's vault permanently
 * @param id {string} - The file/folder ID to delete
 * @param name {string} - The unencrypted name of the content to be deleted
 * @param isFolder {boolean} - True if a folder, else false
 * @param sharedID {string} - The reference ID of the file/folder
 * @param callback {(DeleteResponse)}
 */
const deleteVaultContent = (
    id: string,
    name: string,
    isFolder: boolean,
    sharedID: string,
    callback: (resp: DeleteResponse) => void,
) => {
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
            response.json().then(json => {
                let resp = new interfaces.DeleteResponse(json);
                callback(resp);
            })
        } else {
            alert("Error deleting item");
            showStorageBar("", 0);
        }
    }).catch(error => {
        alert("Error deleting item: " + error);
    });
}

/**
 * Creates a new folder inside the current vault folder
 * @param folderName {string}
 */
const createNewFolder = async (folderName: string) => {
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
            let vaultFolder = new interfaces.VaultFolder();
            vaultFolder.id = response.id;
            vaultFolder.refID = response.id;
            vaultFolder.name = encNameEncoded;
            vaultFolder.modified = new Date();
            vaultFolder.protectedKey = encFolderKey;
            vaultFolder.isOwner = cache[folderID].folder.isOwner;
            vaultFolder.canModify = cache[folderID].folder.canModify;
            vaultFolder.sharedBy = "";
            vaultFolder.sharedWith = 0;

            let vaultViewFolder = new VaultViewFolder(
                vaultFolder,
                newFolderKeyImported,
                folderName);
            currentFolders[response.id] = vaultViewFolder;
            let row = await generateFolderRow(vaultViewFolder);
            cache[folderID].folders.unshift(vaultFolder);
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

/**
 * Updates the vault message below the progress bar
 * @param msg {string}
 */
const setVaultMessage = (msg: string) => {
    let vaultMessage = document.getElementById("vault-message");
    vaultMessage.innerHTML = `<img class="small-icon progress-spinner" src="/static/icons/progress.svg">${msg}`;
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}
