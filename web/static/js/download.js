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
        }
    };

    xhr.send();
});

const handleMetadata = (download) => {
    // Attempt to decrypt without a password first
    let salt = base64ToArray(download.salt);
    let pepper = location.hash.slice(1);

    deriveKey("", salt, pepper, () => {}, (key, _) => {
        let decryptedName = decryptName(key, download.name);

        if (decryptedName) {
            showDownload(decryptedName, download, key);
        } else {
            promptPassword(download);
        }
    });
}

const showDownload = (name, download, key) => {
    let loading = document.getElementById("loading");
    loading.style.display = "none";

    let passwordPrompt = document.getElementById("password-prompt-div");
    passwordPrompt.style.display = "none";

    let nameSpan = document.getElementById("name");
    nameSpan.textContent = name;

    let downloadDiv = document.getElementById("download-prompt-div");
    downloadDiv.style.display = "inherit";

    let downloadBtn = document.getElementById("download-nopass");
    downloadBtn.addEventListener("click", () => {
        downloadFile(name, download, key);
    })
}

const decryptName = (key, name) => {
    let nameBytes = hexToBytes(name);
    name = decryptString(key, nameBytes);
    return name;
}

const updatePasswordBtn = (txt, disabled) => {
    let btn = document.getElementById("submit");
    btn.value = txt;
    btn.textContent = txt;
    btn.disabled = disabled;
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
            updatePasswordBtn("Validating", true);
        }, (key, _) => {
            let decryptedName = decryptName(key, download.name);

            if (decryptedName) {
                showDownload(decryptedName, download, key);
            } else {
                updatePasswordBtn("Submit", false);
                alert("Incorrect password");
            }
        });
    });
}

const downloadFile = (name, download, key) => {
    let writer = getFileWriter(name);

    const fetch = (chunkNum) => {
        let xhr = new XMLHttpRequest();
        xhr.open("GET", `/d/${download.id}/${chunkNum}`, true);
        xhr.responseType = 'blob';

        xhr.onreadystatechange = async () => {
            if (xhr.readyState === 4 && xhr.status === 200) {
                let data = new Uint8Array(await xhr.response.arrayBuffer());
                let decryptedChunk = decryptChunk(key, data);
                writer.write(decryptedChunk).then(() => {
                    if (chunkNum === download.chunks) {
                        writer.close().then(r => console.log(r));
                    } else {
                        // Fetch next chunk
                        fetch(chunkNum + 1);
                    }
                });
            } else if (xhr.readyState === 4 && xhr.status !== 200) {
                alert(`Error ${xhr.status}: ${xhr.responseText}`);
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

    let fileStream = streamSaver.createWriteStream(name);
    return fileStream.getWriter();
}