const streamSaver = window.streamSaver;

document.addEventListener("DOMContentLoaded", () => {
    let xhr = new XMLHttpRequest();
    xhr.open("GET", `/d${window.location.pathname}`, true);
    xhr.setRequestHeader('Content-Type', 'application/json');

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            let download = JSON.parse(xhr.responseText);
            handleMetadata(download);
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            alert(`Error ${xhr.status}: ${xhr.responseText}`);
            window.location = "/";
        }
    };

    xhr.send();
});

const handleMetadata = (download) => {
    // Attempt to decrypt without a password first
    let salt = base64ToArray(download.salt);
    let pepper = location.hash.slice(1);

    deriveKey("", salt, pepper, () => {}, async (key, _) => {
        decryptName(key, download.name).then(result => {
            showDownload(result, download, key);
        }).catch(err => {
            console.warn(err);
            promptPassword(download);
        });
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

    let downloadBtn = document.getElementById("download-nopass");
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
            downloadFile(name, download, key, downloadCallback);
        }
    })
}

const decryptName = async (key, name) => {
    let nameBytes = hexToBytes(name);
    return await decryptString(key, nameBytes);
}

const updatePasswordBtn = (txt, disabled) => {
    let btn = document.getElementById("submit");
    btn.value = txt;
    btn.textContent = txt;
    btn.disabled = disabled;
}

const setFormEnabled = on => {
    let fieldset = document.getElementById("download-fieldset");
    fieldset.disabled = !on;
}

const calcTimeRemaining = expiration => {
    let currentTime = new Date();
    let expTime = new Date(expiration);

    let timeDifference = expTime - currentTime;

    const totalSeconds = Math.floor(timeDifference / 1000);
    const totalMinutes = Math.floor(totalSeconds / 60);
    const totalHours = Math.floor(totalMinutes / 60);
    const days = Math.floor(totalHours / 24);

    const hours = totalHours % 24;
    const minutes = totalMinutes % 60;
    const seconds = totalSeconds % 60;

    return `${days} days, ${hours} hours, ${minutes} minutes, ${seconds} seconds`;
}

const calcFileSize = bytes => {
    let thresh = 1000;

    if (Math.abs(bytes) < thresh) {
        return bytes + ' B';
    }

    const units = ['KB', 'MB', 'GB', 'TB'];
    let u = -1;
    const r = 10;

    do {
        bytes /= thresh;
        ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);


    return bytes.toFixed(1) + ' ' + units[u];
}

const promptPassword = (download) => {
    let loading = document.getElementById("loading");
    loading.style.display = "none";

    let downloadDiv = document.getElementById("password-prompt-div");
    downloadDiv.style.display = "inherit";

    let password = document.getElementById("password");
    let btn = document.getElementById("submit");

    btn.addEventListener("click", () => {
        let salt = base64ToArray(download.salt);
        let pepper = location.hash.slice(1);

        deriveKey(password.value, salt, pepper, () => {
            setFormEnabled(false);
            updatePasswordBtn("Validating", true);
        }, async (key, _) => {
            setFormEnabled(true);

            let decryptedName = await decryptName(key, download.name);

            if (decryptedName) {
                showDownload(decryptedName, download, key);
            } else {
                updatePasswordBtn("Submit", false);
                alert("Incorrect password");
            }
        });
    });
}

const downloadText = (download, key, callback) => {
    const fetch = () => {
        let xhr = new XMLHttpRequest();
        xhr.open("GET", `/d/${download.id}/1`, true);
        xhr.responseType = "blob";

        xhr.onreadystatechange = async () => {
            if (xhr.readyState === 4 && xhr.status === 200) {
                let data = new Uint8Array(await xhr.response.arrayBuffer());
                let decryptedChunk = await decryptChunk(key, data);
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

const downloadFile = (name, download, key, callback) => {
    let writer = getFileWriter(name);

    const fetch = (chunkNum) => {
        let xhr = new XMLHttpRequest();
        xhr.open("GET", `/d/${download.id}/${chunkNum}`, true);
        xhr.responseType = "blob";

        xhr.onreadystatechange = async () => {
            if (xhr.readyState === 4 && xhr.status === 200) {
                let data = new Uint8Array(await xhr.response.arrayBuffer());
                let decryptedChunk = await decryptChunk(key, data);
                writer.write(new Uint8Array(decryptedChunk)).then(() => {
                    if (chunkNum === download.chunks) {
                        writer.close().then(r => console.log(r));
                        callback(true);
                    } else {
                        // Fetch next chunk
                        fetch(chunkNum + 1);
                    }
                });
            } else if (xhr.readyState === 4 && xhr.status !== 200) {
                alert(`Error ${xhr.status}: ${xhr.responseText}`);
                callback(false);
            }
        };

        xhr.send();
    }

    // Start with first chunk
    fetch(1);
}

const getFileWriter = (name, length) => {
    // TODO: Need original file size sent to and received from server
    // let fileStream = streamSaver.createWriteStream(name, {
    //     size: length, // (optional filesize) Will show progress
    //     writableStrategy: undefined, // (optional)
    //     readableStrategy: undefined  // (optional)
    // });

    // StreamSaver's "mitm" technique for downloading large files only works
    // over https. If served over http, it'll default to:
    // https://jimmywarting.github.io/StreamSaver.js/mitm.html?version=2.0.0
    if (location.protocol.includes("https")) {
        window.streamSaver.mitm = "/mitm.html";
    }

    let fileStream = window.streamSaver.createWriteStream(name);
    return fileStream.getWriter();
}