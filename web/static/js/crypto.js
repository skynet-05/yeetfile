const HashSize = 32;
const IVSize = 12;
export const TotalOverhead = 28;
let utf8Encode = new TextEncoder();
let utf8Decode = new TextDecoder();

let webcrypto;
let indexedDB;

/**
 * deriveSendingKey creates a PBKDF2 key using a password as the payload, a
 * salt (or a randomly generated salt if not provided) and a pepper to append
 * to the password. Returns the derived key and the salt.
 * @param password {string} - the password for generating the key
 * @param salt {Uint8Array} - the key salt (can be left undefined to randomly generate one)
 * @param pepper {string} - the pepper to extend the password
 * @returns {Promise<[CryptoKey,Uint8Array]>}
 */
export const deriveSendingKey = async (password, salt, pepper) => {
    if (!salt) {
        salt = webcrypto.getRandomValues(new Uint8Array(HashSize));
    }

    password = password + pepper;
    let encodedPassword = utf8Encode.encode(password);

    return [await deriveKey(encodedPassword, salt), salt];
}

/**
 * importKey imports Uint8Array key data as a CryptoKey object
 * @param keyData {Uint8Array}
 * @returns {Promise<CryptoKey>}
 */
export const importKey = async (keyData) => {
    return await webcrypto.subtle.importKey(
        "raw",
        keyData,
        "AES-GCM",
        false,
        ["encrypt", "decrypt"]
    )
}

/**
 * deriveKey derives a PBKDF2 key from a password and salt
 * @param password {Uint8Array} - a UTF-8 encoded password for the key
 * @param salt {Uint8Array} - the salt for the key
 * @returns {Promise<CryptoKey>}
 */
export const deriveKey = async (password, salt) => {
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
 * @param key {CryptoKey} - a PBKDF2 key
 * @param str {string} - the string to encrypt
 * @returns {Promise<Uint8Array>}
 */
export const encryptString = async (key, str) => {
    let data = utf8Encode.encode(str);
    return await encryptChunk(key, data);
}

/**
 * exportKey exports a PBKDF2 key to a Uint8Array
 * @param key {CryptoKey} - the PBKDF2 key
 * @param format {string} - the format to use when exporting the key (default "raw")
 * @returns {Promise<Uint8Array>}
 */
export const exportKey = async (key, format) => {
    const exported = await webcrypto.subtle.exportKey(format ? format : "raw", key);
    return new Uint8Array(exported);
}

/**
 * encryptChunk encrypts a chunk of data using the provided PBKDF2 key, and
 * returns the encrypted chunk with the initialization vector prepended to the
 * encrypted data.
 * @param key {CryptoKey} - the PBKDF2 key
 * @param data {Uint8Array} - the data to encrypt
 * @returns {Promise<Uint8Array>}
 */
export const encryptChunk = async (key, data) => {
    let iv = webcrypto.getRandomValues(new Uint8Array(IVSize));
    let encrypted = await webcrypto.subtle.encrypt({ name: "AES-GCM", iv }, key, data);
    let merged = new Uint8Array(iv.length + encrypted.byteLength);
    merged.set(iv);
    merged.set(new Uint8Array(encrypted), iv.length);

    return merged;
}

/**
 * encryptRSA encrypts a Uint8Array using an RSA-OAEP public key
 * @param key {CryptoKey} - the RSA-OAEP public key
 * @param data {Uint8Array} - the data to encrypt
 * @returns {Promise<Uint8Array>}
 */
export const encryptRSA = async (key, data) => {
    let encrypted = await webcrypto.subtle.encrypt({ name: "RSA-OAEP" }, key, data);
    return new Uint8Array(encrypted);
}

/**
 * decryptRSA decrypts an encrypted Uint8Array using an RSA-OAEP private key
 * @param key {CryptoKey} - the RSA-OAEP private key
 * @param data {Uint8Array} - the data to decrypt
 * @returns {Promise<Uint8Array>}
 */
export const decryptRSA = async (key, data) => {
    let decrypted = await webcrypto.subtle.decrypt({ name: "RSA-OAEP" }, key, data);
    return new Uint8Array(decrypted);
}

/**
 * decryptString decrypts an encrypted string using the provided key
 * @param key {CryptoKey} - the PBKDF2 key to use for decryption
 * @param data {Uint8Array} - the encrypted string data to decrypt
 * @returns {Promise<string>}
 */
export const decryptString = async (key, data) => {
    let str = await decryptChunk(key, data);
    return utf8Decode.decode(str);
}

/**
 * decryptChunk decrypts a chunk of AES-GCM 256 encrypted data using
 * the provided key
 * @param key {CryptoKey} - the key to use for decryption
 * @param data {Uint8Array} - the encrypted data to decrypt
 * @returns {Promise<ArrayBuffer>}
 */
export const decryptChunk = async (key, data) => {
    let iv = data.slice(0, IVSize);
    let fileData = data.slice(IVSize, data.length + 1);

    return await webcrypto.subtle.decrypt({ name: "AES-GCM", iv }, key, fileData);
}

/**
 * generateUserKey creates a PBKDF2 key using user's password as the payload and
 * their identifier (email or account ID) as the salt.
 * @param identifier {string} - the user's email or account ID
 * @param password {string} - the user's password
 * @returns {Promise<CryptoKey>}
 */
export const generateUserKey = async (identifier, password) => {
    return await deriveKey(utf8Encode.encode(password), utf8Encode.encode(identifier));
}

/**
 * generateLoginKeyHash generates a user's "login key" a PBKDF2 where the payload is
 * the user's user key and the salt is the user's password, and returns a SHA-256 hash
 * of that login key.
 * @param userKey {CryptoKey} - the user's user key from generateUserKey
 * @param password {string} - the user's password
 * @returns {Promise<Uint8Array>}
 */
export const generateLoginKeyHash = async (userKey, password) => {
    let userKeyExported = await exportKey(userKey, "raw");

    let loginKey = await deriveKey(userKeyExported, utf8Encode.encode(password));
    let loginKeyExported = await exportKey(loginKey, "raw");
    let loginKeyHash = await webcrypto.subtle.digest("SHA-256", loginKeyExported);

    return new Uint8Array(loginKeyHash);
}

/**
 * generateRandomKey uses a CSPRNG to generate a 32-byte key for encrypting and
 * decrypting folder contents in YeetFile. This is always encrypted with the user's
 * public key before being sent to the server.
 * @returns {Uint8Array}
 */
export const generateRandomKey = () => {
    return webcrypto.getRandomValues(new Uint8Array(HashSize));
}

/**
 * fetchWordlist requests the list of words from the server, which is used
 * for generating random passphrases as peppers for files.
 * @param callback - the request callback
 */
export const fetchWordlist = callback => {
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
 * @param callback {function}
 */
export const generatePassphrase = (callback) => {
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

/**
 * ingestPublicKey takes the raw base64 of the user's public key and
 * converts them into a CryptoKey object that can be used for encryption.
 * @param publicKey {string}
 * @param callback {function(CryptoKey)}
 */
export const ingestPublicKey = (publicKey, callback) => {
    let decodedPublicKey = base64ToArray(publicKey);

    webcrypto.subtle.importKey(
        "spki",
        decodedPublicKey,
        {
            name: "RSA-OAEP",
            hash: {name: "SHA-256"}
        },
        false,
        ["encrypt"]
    ).catch(error => {
        console.error("Error re-importing vault key:", error);
    }).then(key => {
        callback(key);
    });
}

/**
 * ingestProtectedKey uses the user key to decrypt the protected key, and
 * uses the result to create a non-exportable CryptoKey object that is then
 * stored in IndexedDB.
 * @param userKey {CryptoKey}
 * @param protectedKey {Uint8Array}
 * @param callback {function(CryptoKey)}
 */
export const ingestProtectedKey = (userKey, protectedKey, callback) => {
    let decodedProtectedKey = base64ToArray(protectedKey);
    let iv = decodedProtectedKey.slice(0, IVSize);

    decodedProtectedKey = decodedProtectedKey.slice(IVSize, decodedProtectedKey.length + 1);

    webcrypto.subtle.decrypt(
        {
            name: "AES-GCM",
            iv: iv,
        },
        userKey,
        decodedProtectedKey
    )
        .then(decryptedData => {
            // Re-import the key as non-exportable
            webcrypto.subtle.importKey(
                "pkcs8",
                decryptedData,
                {
                    name: "RSA-OAEP",
                    hash: {name: "SHA-256"}
                },
                false, // Key cannot be exported
                ["decrypt"]
            )
                .catch(error => {
                    console.error("Error re-importing vault key:", error);
                })
                .then(key => {
                    callback(key);
                });
        })
        .catch(error => {
            alert("Error decrypting vault key");
            console.error("Error decrypting vault key:", error);
        });
}

/**
 * generateKeyPair generates RSA-OAEP public + private keys. The private key
 * is used for encrypting/decrypting the user's root folder, as well as folder
 * keys that are shared with the user. The public key is used by other users to
 * share folders.
 *
 * Note that the generated key pair is marked as "extractable", since the private
 * key must be further encrypted by the user key before being sent to the server.
 *
 * @returns {Promise<CryptoKeyPair>}
 */
export const generateKeyPair = async () => {
   return await webcrypto.subtle.generateKey(
        {
            name: 'RSA-OAEP',
            modulusLength: 2048,
            publicExponent: new Uint8Array([0x01, 0x00, 0x01]), // 65537
            hash: {name: 'SHA-256'}
        }, true, ['encrypt', 'decrypt']
    );
}

if (typeof window === 'undefined') {
    // Running in Node.js
    webcrypto = require('crypto').webcrypto;

    module.exports = {
        deriveSendingKey,
        encryptString,
        encryptChunk,
        encryptRSA,
        decryptRSA,
        generatePassphrase,
        generateUserKey,
        generateLoginKeyHash,
        generateRandomKey: generateRandomKey,
        fetchWordlist,
        decryptChunk,
        decryptString,
        exportKey,
        importKey,
        generateKeyPair,
        webcrypto
    };
} else {
    // Running in a browser
    webcrypto = window.crypto;
    indexedDB = window.indexedDB || window.mozIndexedDB || window.webkitIndexedDB || window.msIndexedDB || window.shimIndexedDB
}