const HashSize = 32;
const NonceSize = 16;
let utf8Encode = new TextEncoder();
let utf8Decode = new TextDecoder();

let webcrypto;

/**
 * deriveSendingKey creates a PBKDF2 key using a password as the payload, a
 * salt (or a randomly generated salt if not provided) and a pepper to append
 * to the password. Returns the derived key and the salt.
 * @param password - the password for generating the key
 * @param salt - the key salt (can be left undefined to randomly generate one)
 * @param pepper - the pepper to extend the password
 * @returns {Promise<[CryptoKey,Uint8Array]>}
 */
const deriveSendingKey = async (password, salt, pepper) => {
    if (!salt) {
        salt = webcrypto.getRandomValues(new Uint8Array(HashSize));
    }

    password = password + pepper;
    password = utf8Encode.encode(password);

    return [await deriveKey(password, salt), salt];
}

/**
 * deriveKey derives a PBKDF2 key from a password and salt
 * @param password - a Uint8Array password for the key
 * @param salt - the Uint8Array salt for the key
 * @returns {Promise<CryptoKey>}
 */
const deriveKey = async (password, salt) => {
    let keyMaterial = await webcrypto.subtle.importKey(
        "raw",
        password,
        "PBKDF2",
        false,
        ["deriveBits", "deriveKey"],
    );

    return await webcrypto.subtle.deriveKey(
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
    );
}

/**
 * encryptString encrypts a string `str` using the PBKDF2 key `key` using
 * AES-GCM 256 encryption.
 * @param key - a PBKDF2 key
 * @param str - the string to encrypt
 * @returns {Promise<Uint8Array>}
 */
const encryptString = async (key, str) => {
    let data = utf8Encode.encode(str);
    return await encryptChunk(key, data);
}

/**
 * exportKey exports a PBKDF2 key to a Uint8Array
 * @param key - the PBKDF2 key
 * @returns {Promise<Uint8Array>}
 */
const exportKey = async (key) => {
    const exported = await webcrypto.subtle.exportKey("raw", key);
    return new Uint8Array(exported);
}

/**
 * encryptChunk encrypts a chunk of data using the provided PBKDF2 key, and
 * returns the encrypted chunk with the initialization vector prepended to the
 * encrypted data.
 * @param key - the PBKDF2 key
 * @param data - the data to encrypt
 * @returns {Promise<Uint8Array>}
 */
const encryptChunk = async (key, data) => {
    let iv = webcrypto.getRandomValues(new Uint8Array(NonceSize));

    let encrypted = await webcrypto.subtle.encrypt({ name: "AES-GCM", iv }, key, data);
    let merged = new Uint8Array(iv.length + encrypted.byteLength);
    merged.set(iv);
    merged.set(new Uint8Array(encrypted), iv.length);

    return merged;
}

/**
 * decryptString decrypts an encrypted string using the provided key
 * @param key - the PBKDF2 key to use for decryption
 * @param data - the encrypted string data to decrypt
 * @returns {Promise<string>}
 */
const decryptString = async (key, data) => {
    let str = await decryptChunk(key, data);
    return utf8Decode.decode(str);
}

/**
 * decryptChunk decrypts a chunk of AES-GCM 256 encrypted data using
 * the provided PBKDF2 key
 * @param key - the PBKDF2 key to use for decryption
 * @param data - the encrypted data to decrypt
 * @returns {Promise<ArrayBuffer>}
 */
const decryptChunk = async (key, data) => {
    let iv = data.slice(0, NonceSize);
    data = data.slice(NonceSize, data.length + 1);

    return await webcrypto.subtle.decrypt({ name: "AES-GCM", iv }, key, data);
}

/**
 * generateUserKey creates a PBKDF2 key using user's password as the payload and
 * their identifier (email or account ID) as the salt.
 * @param identifier - the user's email or account ID
 * @param password - the user's password
 * @returns {Promise<CryptoKey>}
 */
const generateUserKey = async (identifier, password) => {
    return await deriveKey(utf8Encode.encode(password), utf8Encode.encode(identifier));
}

/**
 * generateLoginKeyHash generates a user's "login key" a PBKDF2 where the payload is
 * the user's user key and the salt is the user's password, and returns a SHA-256 hash
 * of that login key.
 * @param userKey - the user's user key from generateUserKey
 * @param password - the user's password
 * @returns {Promise<Uint8Array>}
 */
const generateLoginKeyHash = async (userKey, password) => {
    let userKeyExported = await exportKey(userKey);

    let loginKey = await deriveKey(userKeyExported, utf8Encode.encode(password));
    let loginKeyExported = await exportKey(loginKey);
    let loginKeyHash = await webcrypto.subtle.digest("SHA-256", loginKeyExported);

    return new Uint8Array(loginKeyHash);
}

/**
 * generateStorageKey uses a CSPRNG to generate a 32-byte key for encrypting and
 * decrypting files stored in YeetFile. This is always encrypted with the user key
 * before being sent to the server.
 * @returns {Uint8Array}
 */
const generateStorageKey = () => {
    return webcrypto.getRandomValues(new Uint8Array(HashSize));
}

/**
 * fetchWordlist requests the list of words from the server, which is used
 * for generating random passphrases as peppers for files.
 * @param callback - the request callback
 */
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

/**
 * generatePassphrase generates a unique 3 word passphrase separated by
 * "." and including a random number in a random position.
 * @param callback
 */
const generatePassphrase = (callback) => {
    let passphrase = [];
    fetchWordlist(words => {
        let wordNum = Math.floor(Math.random() * 3);
        let randNum = Math.floor(Math.random() * 10);
        let numBefore = Math.round(Math.random());

        while (passphrase.length < 3) {
            let idx = Math.floor(Math.random() * (words.length + 1));
            let word = words[idx];
            if (wordNum === 0) {
                if (numBefore) {
                    word = randNum + word;
                } else {
                    word = word + randNum;
                }
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