const HashSize = 32;
const NonceSize = 16;
let utf8Encode = new TextEncoder();
let utf8Decode = new TextDecoder();

let webcrypto;

const deriveSendingKey = (password, salt, pepper, updateCallback, keyCallback) => {
    if (!salt) {
        salt = webcrypto.getRandomValues(new Uint8Array(HashSize));
    }

    password = password + pepper;

    deriveKey(password, salt, updateCallback, keyCallback);
}

const deriveKey = (password, salt, updateCallback, keyCallback) => {
    webcrypto.subtle.importKey(
        "raw",
        utf8Encode.encode(password),
        "PBKDF2",
        false,
        ["deriveBits", "deriveKey"],
    ).then((keyMaterial) => {
        webcrypto.subtle.deriveKey(
            {
                name: "PBKDF2",
                salt,
                iterations: 600000,
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

const exportKey = async (key) => {
    const exported = await webcrypto.subtle.exportKey("raw", key);
    return new Uint8Array(exported);
}

const encryptChunk = async (key, data) => {
    let iv = webcrypto.getRandomValues(new Uint8Array(NonceSize));

    let encrypted = await webcrypto.subtle.encrypt({ name: "AES-GCM", iv }, key, data);
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

    return await webcrypto.subtle.decrypt({ name: "AES-GCM", iv }, key, data);
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

if (typeof window === 'undefined') {
    // Running in Node.js
    webcrypto = require('crypto').webcrypto;

    module.exports = {
        deriveSendingKey,
        encryptString,
        encryptChunk,
        generatePassphrase,
        fetchWordlist,
        decryptChunk,
        decryptString,
        exportKey,
        webcrypto
    };
} else {
    // Running in a browser
    webcrypto = window.crypto;
}