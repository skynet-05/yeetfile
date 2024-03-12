const HashSize = 32;
const NonceSize = 12;
let utf8Encode = new TextEncoder();
let utf8Decode = new TextDecoder();

const deriveKey = (password, salt, pepper, updateCallback, keyCallback) => {
    if (!salt) {
        salt = window.crypto.getRandomValues(new Uint8Array(HashSize));
    }

    password = password + pepper;

    window.crypto.subtle.importKey(
        "raw",
        utf8Encode.encode(password),
        "PBKDF2",
        false,
        ["deriveBits", "deriveKey"],
    ).then((keyMaterial) => {
        window.crypto.subtle.deriveKey(
            {
                name: "PBKDF2",
                salt,
                iterations: 100000,
                hash: "SHA-256",
            },
            keyMaterial,
            { name: "AES-GCM", length: 256 },
            true,
            ["encrypt", "decrypt"],
        ).then((key) => {
            keyCallback(key, salt);
        });
    });
}

const encryptString = async (key, str) => {
    let data = utf8Encode.encode(str);
    return await encryptChunk(key, data);
}

const encryptChunk = async (key, data) => {
    let iv = window.crypto.getRandomValues(new Uint8Array(NonceSize));

    let encrypted = await window.crypto.subtle.encrypt({ name: "AES-GCM", iv }, key, data);
    let merged = new Uint8Array(iv.length + encrypted.byteLength);
    merged.set(iv);
    merged.set(new Uint8Array(encrypted), iv.length);

    return merged;
}

const decryptString = async (key, data) => {
    let str = await decryptChunk(key, data);
    return utf8Decode.decode(str);
}

const decryptChunk = async (key, data) => {
    let iv = data.slice(0, NonceSize);
    data = data.slice(NonceSize, data.length + 1);

    return await window.crypto.subtle.decrypt({ name: "AES-GCM", iv }, key, data);
}

const fetchWordlist = callback => {
    fetch("/wordlist")
        .then((response) => response.json())
        .then((data) => {
            callback(data);
        })
        .catch((error) => {
            console.error("Error fetching wordlist: ", error);
        });
}

const generatePassphrase = (callback) => {
    let passphrase = [];
    fetchWordlist(words => {
        let wordNum = Math.floor(Math.random() * 3);
        let randNum = Math.floor(Math.random() * 10);

        while (passphrase.length < 3) {
            let idx = Math.floor(Math.random() * (words.length + 1));
            let word = words[idx];
            if (wordNum === 0) {
                word += randNum;
            }
            wordNum--;
            passphrase.push(word);
        }

        callback(passphrase.join("."));
    })
}