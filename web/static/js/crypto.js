const HashSize = 32;
const NonceSize = 24;
let utf8Encode = new TextEncoder();
let utf8Decode = new TextDecoder();

const deriveKey = (password, salt, pepper, updateCallback, keyCallback) => {
    if (!salt) {
        salt = nacl.randomBytes(HashSize);
    }

    password = utf8Encode.encode(password + pepper);
    console.log(password);
    console.log(pepper);

    let keyPromise = scrypt.scrypt(password, salt, 32768, 8, 1, HashSize, updateCallback);
    keyPromise.then((key) => {
        keyCallback(key, salt)
    });
}

const encryptString = (key, str) => {
    let data = utf8Encode.encode(str);
    return encryptChunk(key, data);
}

const encryptChunk = (key, data) => {
    let nonce = nacl.randomBytes(NonceSize);

    let encryptedData = nacl.secretbox(data, nonce, key);
    let merged = new Uint8Array(nonce.length + encryptedData.length);
    merged.set(nonce);
    merged.set(encryptedData, nonce.length);

    return merged;
}

const decryptString = (key, data) => {
    let str = decryptChunk(key, data);
    return utf8Decode.decode(str);

}

const decryptChunk = (key, data) => {
    let nonce = data.slice(0, NonceSize)
    data = data.slice(NonceSize, data.length + 1);

    return nacl.secretbox.open(data, nonce, key);
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