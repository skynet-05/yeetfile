import * as crypto from "./crypto.js";
import * as transfer from "./transfer.js";
import {Endpoints} from "./endpoints.js";

const init = () => {
    let xhr = new XMLHttpRequest();
    let id = window.location.pathname.split("/").slice(-1)[0];
    console.log(id);
    let url = Endpoints.format(Endpoints.DownloadSendFileMetadata, id);
    xhr.open("GET", url, true);
    xhr.setRequestHeader('Content-Type', 'application/json');

    xhr.onreadystatechange = async () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            let download = JSON.parse(xhr.responseText);
            await handleMetadata(download);
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            alert(`Error ${xhr.status}: ${xhr.responseText}`);
            window.location.assign("/");
        }
    };

    xhr.send();
};

const handleMetadata = async (download) => {
    // Attempt to decrypt without a password first
    let salt = base64ToArray(download.salt);
    let pepper = location.hash.slice(1);

    let [key, _] = await crypto.deriveSendingKey("", salt, pepper)
    decryptName(key, download.name).then(result => {
        showDownload(result, download, key);
    }).catch(err => {
        console.warn(err);
        promptPassword(download);
    });
}

const showDownload = (name, download, key) => {
    let loading = document.getElementById("loading");
    loading.style.display = "none";

    let passwordPrompt = document.getElementById("password-prompt-div");
    passwordPrompt.style.display = "none";

    let nameSpan = document.getElementById("name");
    if (download.id.startsWith("text_")) {
        nameSpan.textContent = "N/A (text-only)";
    } else {
        nameSpan.textContent = name;
    }

    let expiration = document.getElementById("expiration");
    expiration.textContent = calcTimeRemaining(download.expiration);

    let downloads = document.getElementById("downloads");
    downloads.textContent = download.downloads;

    let size = document.getElementById("size");
    size.textContent = calcFileSize(download.size);

    let downloadDiv = document.getElementById("download-prompt-div");
    downloadDiv.style.display = "inherit";

    let downloadBtn = document.getElementById("download-nopass") as HTMLButtonElement;
    downloadBtn.addEventListener("click", () => {
        downloadBtn.disabled = true;
        downloadBtn.innerText = "Downloading...";

        const downloadCallback = success => {
            if (success) {
                let newDownloadCount = parseInt(download.downloads) - 1;
                downloadBtn.style.display = "none";
                downloads.textContent = String(newDownloadCount);

                if (newDownloadCount === 0) {
                    downloads.textContent += " (deleted)";
                }
            } else {
                downloadBtn.disabled = false;
            }
        }

        if (download.id.startsWith("text_")) {
            downloadText(download, key, downloadCallback);
        } else {
            transfer.downloadSentFile(name, download, key, downloadCallback, () => {});
        }
    })
}

const decryptName = async (key, name) => {
    let nameBytes = hexToBytes(name);
    return await crypto.decryptString(key, nameBytes);
}

const updatePasswordBtn = (txt, disabled) => {
    let btn = document.getElementById("submit") as HTMLButtonElement;
    btn.value = txt;
    btn.textContent = txt;
    btn.disabled = disabled;
}

const setFormEnabled = on => {
    let fieldset = document.getElementById("download-fieldset") as HTMLFieldSetElement;
    fieldset.disabled = !on;
}

const calcTimeRemaining = expiration => {
    let currentTime = new Date();
    let expTime = new Date(expiration);

    let timeDifference = expTime.getTime() - currentTime.getTime();

    const totalSeconds = Math.floor(timeDifference / 1000);
    const totalMinutes = Math.floor(totalSeconds / 60);
    const totalHours = Math.floor(totalMinutes / 60);
    const days = Math.floor(totalHours / 24);

    const hours = totalHours % 24;
    const minutes = totalMinutes % 60;
    const seconds = totalSeconds % 60;

    return `${days} days, ${hours} hours, ${minutes} minutes, ${seconds} seconds`;
}

const promptPassword = (download) => {
    let loading = document.getElementById("loading");
    loading.style.display = "none";

    let downloadDiv = document.getElementById("password-prompt-div");
    downloadDiv.style.display = "inherit";

    let password = document.getElementById("password") as HTMLInputElement;
    let btn = document.getElementById("submit") as HTMLButtonElement;

    btn.addEventListener("click", async () => {
        let salt = base64ToArray(download.salt);
        let pepper = location.hash.slice(1);

        setFormEnabled(false);
        updatePasswordBtn("Validating", true);

        let [key, _] = await crypto.deriveSendingKey(password.value, salt, pepper);

        setFormEnabled(true);

        decryptName(key, download.name).then(decryptedName => {
            showDownload(decryptedName, download, key);
        }).catch(() => {
            updatePasswordBtn("Submit", false);
            alert("Incorrect password");
        });
    });
}

const downloadText = (download, key, callback) => {
    const fetch = () => {
        let xhr = new XMLHttpRequest();
        let url = Endpoints.format(Endpoints.DownloadSendFileData, download.id);
        xhr.open("GET", url, true);
        xhr.responseType = "blob";

        xhr.onreadystatechange = async () => {
            if (xhr.readyState === 4 && xhr.status === 200) {
                let data = new Uint8Array(await xhr.response.arrayBuffer());
                let decryptedChunk = await crypto.decryptChunk(key, data);
                let decryptedText = new TextDecoder().decode(decryptedChunk);
                displayText(decryptedText);
                callback(true);
            } else if (xhr.readyState === 4 && xhr.status !== 200) {
                alert(`Error ${xhr.status}: ${xhr.responseText}`);
                callback(false);
            }
        };

        xhr.send();
    }

    fetch();
}

const displayText = (text) => {
    let plaintextDiv = document.getElementById("plaintext-div");
    let plaintextContent = document.getElementById("plaintext-content");

    plaintextDiv.style.display = "initial";
    plaintextContent.innerText = text;
}

if (document.readyState !== "loading") {
    init();
} else {
    document.addEventListener("DOMContentLoaded", () => {
        init();
    });
}