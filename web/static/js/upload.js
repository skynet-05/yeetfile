const expUnits = {
    minutes: 0,
    hours: 1,
    days: 2
}

let pepper = "";

document.addEventListener("DOMContentLoaded", () => {
    setupTypeToggles();

    let usePasswordCB = document.getElementById("use-password");
    let passwordDiv = document.getElementById("password-div");
    usePasswordCB.addEventListener("change", (event) => {
        if (event.currentTarget.checked) {
            passwordDiv.style.display = "inherit";
        } else {
            passwordDiv.style.display = "none";
        }
    });

    let uploadTextContent = document.getElementById("upload-text-content");
    let uploadTextLabel = document.getElementById("upload-text-label");
    uploadTextContent.addEventListener("input", () => {
        if (uploadTextLabel) {
            uploadTextLabel.innerText=`Text (${uploadTextContent.value.length}/1000):`;
        }
    });

    let form = document.getElementById("upload-form");
    let nameDiv = document.getElementById("name-div");
    let filePicker = document.getElementById("upload");
    filePicker.addEventListener("change", () => {
        if (filePicker.files.length > 1) {
            nameDiv.style.display = "inherit";
        } else {
            nameDiv.style.display = "none";
        }
    });

    form.addEventListener("reset", (event) => {
        resetForm();
    });

    form.addEventListener("submit", (event) => {
        event.preventDefault();

        let formValues = getFormValues();

        if (validateForm(formValues)) {
            setFormEnabled(false);
            generatePassphrase(passphrase => {
                pepper = passphrase;
                deriveKey(formValues.pw, undefined, passphrase, () => {
                    updateProgress("Initializing...")
                }, (key, salt) => {
                    if (formValues.files.length > 1) {
                        submitFormMulti(formValues, key, salt, allowReset);
                    } else {
                        submitFormSingle(formValues, key, salt, allowReset);
                    }
                });
            });
        }
    });
});

const setFormEnabled = on => {
    let fieldset = document.getElementById("form-fieldset");
    fieldset.disabled = !on;
}

const updateProgress = (txt) => {
    let uploadBtn = document.getElementById("submit");
    uploadBtn.disabled = true;
    uploadBtn.value = txt;
}

const allowReset = () => {
    updateProgress("Done!")
    let reset = document.getElementById("reset");
    reset.style.display = "inline";
}

const resetForm = () => {
    let uploadBtn = document.getElementById("submit");
    uploadBtn.disabled = false;
    uploadBtn.value = "Upload";

    let reset = document.getElementById("reset");
    reset.style.display = "none";

    setFormEnabled(true);
}

const getFormValues = () => {
    let files = document.getElementById("upload").files;
    let pw = document.getElementById("password").value;
    let pwConfirm = document.getElementById("confirm-password").value;
    let downloads = document.getElementById("downloads").value;
    let exp = document.getElementById("expiration").value;
    let unit = document.getElementById("duration-unit").selectedIndex;

    // If the password checkbox isn't checked, unset password
    let usePassword = document.getElementById("use-password").checked;
    if (!usePassword) {
        pw = pwConfirm = "";
    }

    return { files, pw, pwConfirm, downloads, exp, unit };
}

const validateForm = (form) => {
    let files = form.files;

    if (!files || files.length === 0) {
        alert("Select at least one file to upload");
        return false;
    }

    if (!validatePassword(form.pw, form.pwConfirm)) {
        alert("Passwords don't match");
        return false;
    }

    if (!validateExpiration(form.exp, form.unit)) {
        return false;
    }

    if (!validateDownloads(form.downloads)) {
        return false;
    }

    // All fields have been validated
    return true;
}

const submitFormMulti = (form, key, salt, callback) => {
    let name = document.getElementById("name").value || "download.zip";
    if (name.endsWith(".zip.zip")) {
        name = name.replace(".zip.zip", ".zip");
    } else if (!name.endsWith(".zip")) {
        name = name + ".zip";
    }

    let zip = JSZip();
    let size = 0;

    for (let i = 0; i < form.files.length; i++) {
        let file = form.files[i];

        if (file.webkitRelativePath) {
            zip.file(file.webkitRelativePath, file);
        } else {
            zip.file(file.name, file);
        }

        size += file.size;
    }

    let encryptedName = encryptString(key, name);

    let hexName = toHexString(encryptedName);
    let chunks = getNumChunks(size);
    let expString = getExpString(form.exp, form.unit);

    updateProgress("Uploading file...");
    uploadMetadata(
        hexName,
        chunks,
        salt,
        parseInt(form.downloads),
        expString,
        (id) => {
            uploadZip(id, key, zip, chunks).then(() => {
                callback();
            });
        });
}

const submitFormSingle = (form, key, salt, callback) => {
    let file = form.files[0];
    let encryptedName = encryptString(key, file.name);

    let hexName = toHexString(encryptedName);
    let chunks = getNumChunks(file.size);
    let expString = getExpString(form.exp, form.unit);

    uploadMetadata(
        hexName,
        chunks,
        salt,
        parseInt(form.downloads),
        expString,
        (id) => {
            uploadFileChunks(id, key, file, chunks).then(() => {
                callback();
            });
        });
}

const sendChunk = (blob, id, chunkNum, callback) => {
    let xhr = new XMLHttpRequest();

    xhr.open("POST", `/u/${id}/${chunkNum}`, false);
    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            callback(xhr.responseText);
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            alert(`Error ${xhr.status}: ${xhr.responseText}`);
            throw new Error("Unable to upload chunk!");
        }
    }

    xhr.send(blob);
}

const uploadZip = async (id, key, zip, chunks) => {
    let i = 0;
    let zipData = new Uint8Array(0);

    zip.generateInternalStream({type:"uint8array"}).on ('data', (data, metadata) => {
        zipData = concatTypedArrays(zipData, data);
        if (zipData.length >= chunkSize) {
            let slice = zipData.subarray(0, chunkSize);
            let blob = encryptChunk(key, slice);

            updateProgress(`Uploading file... ${i + 1}/${chunks}`)
            sendChunk(blob, id, i + 1);
            zipData = zipData.subarray(chunkSize, zipData.length);
            i += 1;
        }
    }).on("end", () => {
        if (zipData.length > 0) {
            let blob = encryptChunk(key, zipData);
            updateProgress(`Uploading file... ${i + 1}/${chunks}`)
            sendChunk(blob, id, i + 1, (tag) => {
                showFileTag(tag);
            });
        }
    }).resume();
}

const uploadFileChunks = async (id, key, file, chunks) => {
    for (let i = 0; i < chunks; i++) {
        let start = i * chunkSize;
        let end = (i + 1) * chunkSize;

        if (end > file.size) {
            end = file.size;
        }

        let data = await file.slice(start, end).arrayBuffer();
        let blob = encryptChunk(key, new Uint8Array(data));

        updateProgress(`Uploading file... ${i + 1}/${chunks}`)
        sendChunk(blob, id, i + 1, (tag) => {
            if (tag) {
                showFileTag(tag);
            }
        });
    }
}

const uploadMetadata = (name, chunks, salt, downloads, exp, callback) => {
    let xhr = new XMLHttpRequest();
    xhr.open("POST", "/u", false);
    xhr.setRequestHeader('Content-Type', 'application/json');

    xhr.onreadystatechange = () => {
        if (xhr.readyState === 4 && xhr.status === 200) {
            callback(xhr.responseText);
        } else if (xhr.readyState === 4 && xhr.status !== 200) {
            alert(`Error ${xhr.status}: ${xhr.responseText}`);
        }
    };

    xhr.send(JSON.stringify({
        name: name,
        chunks: chunks,
        salt: Array.from(salt),
        downloads: downloads,
        expiration: exp
    }));
}

const validatePassword = (pwInput, pwConfirm) => {
    return (pwInput.length === 0 || pwConfirm === pwInput);
}

const validateDownloads = (numDownloads) => {
    let maxDownloads = 10;
    if (numDownloads > maxDownloads) {
        alert(`The number of downloads must be between 0-${maxDownloads}.`);
        return false;
    }

    return true;
}

const validateExpiration = (exp, unit) => {
    let maxDays = 30;
    let maxHours = 24 * maxDays;
    let maxMinutes = 60 * maxHours;

    if (unit === expUnits.minutes) {
        if (exp <= 0 || exp > maxMinutes) {
            alert(`Expiration must be between 0-${maxMinutes} minutes`);
            return false;
        }
    }

    if (unit === expUnits.hours) {
        if (exp <= 0 || exp > maxHours) {
            alert(`Expiration must be between 0-${maxHours} hours`);
            return false;
        }
    }

    if (unit === expUnits.days) {
        if (exp <= 0 || exp > maxDays) {
            alert(`Expiration must be between 0-${maxDays} days`);
            return false;
        }
    }

    return true;
}

const showFileTag = (tag) => {
    let tagDiv = document.getElementById("file-tag-div");
    let fileTag = document.getElementById("file-tag");
    let fileLink = document.getElementById("file-link");

    let link = `${window.location.protocol}//${window.location.host}/${tag}#${pepper}`

    tagDiv.style.display = "inherit";
    fileTag.textContent = `${tag}#${pepper}`;
    fileLink.textContent = link;
    fileLink.href = link;
}

const setupTypeToggles = () => {
    let uploadTextBtn = document.getElementById("upload-text-btn");
    let uploadTextRow = document.getElementById("upload-text-row");

    let uploadFileBtn = document.getElementById("upload-file-btn");
    let uploadFileRow = document.getElementById("upload-file-row");

    uploadTextBtn.addEventListener("click", () => {
        uploadTextRow.style.display = "contents";
        uploadFileRow.style.display = "none";
    });

    uploadFileBtn.addEventListener("click", () => {
        uploadTextRow.style.display = "none";
        uploadFileRow.style.display = "contents";
    });
}