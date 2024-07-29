import * as constants from "./constants.js";
import * as crypto from "./crypto.js";

export class YeetFileDB {
    constructor() {
        this.dbName = "yeetfileDB";
        this.dbVersion = 1;
        this.keysObjectStore = "keys";
        this.privateKeyID = 1;
        this.publicKeyID = 2;
        this.passwordProtectedID = 3;

        const request = indexedDB.open(this.dbName, this.dbVersion);

        request.onupgradeneeded = (event) => {
            const db = event.target.result;
            db.createObjectStore(this.keysObjectStore, { keyPath: "id" });
        };

        /**
         * insertKey encrypts the user's private key with a random key and stores
         * the bytes for the private and public keys in indexeddb
         * @param privateKey {Uint8Array}
         * @param publicKey {Uint8Array}
         * @param password {string}
         * @param callback {function(boolean)}
         */
        this.insertVaultKeyPair = async (privateKey, publicKey, password, callback) => {
            this.removeKeys(() => {});

            let encKey;
            if (password.length > 0) {
                let [vaultEncKey, _] = await crypto.deriveSendingKey(
                    password,
                    hexToBytes("YEETFILE_RANDOM_SESSION_KEY"), "");
                encKey = vaultEncKey;
            } else {
                encKey = await crypto.importKey(hexToBytes("YEETFILE_RANDOM_SESSION_KEY"))
            }

            // Replaced w/ random value on each request (should be cached by browser)
            let encPrivKey = await crypto.encryptChunk(encKey, privateKey);

            let request = indexedDB.open(this.dbName, this.dbVersion);
            request.onsuccess = async (event) => {
                let db = event.target.result;

                let transaction = db.transaction([this.keysObjectStore], "readwrite");
                let objectStore = transaction.objectStore(this.keysObjectStore);
                try {
                    let putPrivateKeyRequest = await objectStore.put({
                        id: this.privateKeyID,
                        key: encPrivKey
                    });
                    putPrivateKeyRequest.onerror = (event) => {
                        console.error("Error storing private key:", event.target.error);
                        callback(false);
                    };

                    let putPublicKeyRequest = await objectStore.put({
                        id: this.publicKeyID,
                        key: publicKey
                    });
                    putPublicKeyRequest.onerror = (event) => {
                        console.error("Error storing public key:", event.target.error);
                        callback(false);
                    };

                    let putPasswordProtectedRequest = await objectStore.put({
                        id: this.passwordProtectedID,
                        key: password.length > 0
                    });
                    putPasswordProtectedRequest.onerror = (event) => {
                        console.error("Error storing password protection flag:",
                            event.target.error);
                        callback(false);
                    };
                } catch (error) {
                    console.error("Error during put operations:", error);
                }

                transaction.onerror = (event) => {
                    console.error("Error adding vault keys to IndexedDB:", event.target.error);
                    alert("Error preparing vault keys");
                };

                transaction.oncomplete = () => {
                    db.close();
                    callback(true);
                };
            }

            request.onerror = () => {
                console.error("Error opening local db");
            }
        }

        /**
         * isPasswordProtected returns whether the current vault keys were
         * encrypted with a user-provided vault password.
         * @param callback {function(boolean)}
         */
        this.isPasswordProtected = (callback) => {
            let request = indexedDB.open(this.dbName, this.dbVersion);
            request.onsuccess = (event) => {
                let db = event.target.result;
                let transaction = db.transaction([this.keysObjectStore], "readonly");
                let objectStore = transaction.objectStore(this.keysObjectStore);
                let passwordProtectedRequest = objectStore.get(this.passwordProtectedID);

                passwordProtectedRequest.onsuccess = (event) => {
                    callback(event.target.result.key);
                }

                passwordProtectedRequest.onerror = () => {
                    alert("Error checking for vault key password");
                    callback(false);
                }

                transaction.onerror = () => {
                    alert("Error checking for vault key password");
                    callback(false);
                }
            }
        }

        /**
         * getVaultKey returns the vault key from the indexeddb, if it's available
         * @param password {string}
         * @param callback {function(CryptoKey, CryptoKey)}
         * @param errorCallback {function()}
         */
        this.getVaultKeyPair = (password, callback, errorCallback) => {
            let request = indexedDB.open(this.dbName, this.dbVersion);

            request.onsuccess = async (event) => {
                let db = event.target.result;

                let decKey;
                if (password.length > 0) {
                    let [vaultDecKey, _] = await crypto.deriveSendingKey(
                        password,
                        hexToBytes("YEETFILE_RANDOM_SESSION_KEY"), "");
                    decKey = vaultDecKey;
                } else {
                    decKey = await crypto.importKey(hexToBytes("YEETFILE_RANDOM_SESSION_KEY"))
                }

                let transaction = db.transaction([this.keysObjectStore], "readonly");
                let objectStore = transaction.objectStore(this.keysObjectStore);
                let privateKeyRequest = objectStore.get(this.privateKeyID);
                let publicKeyRequest = objectStore.get(this.publicKeyID);

                privateKeyRequest.onsuccess = async (event) => {
                    let privateKeyBytes = event.target.result.key;

                    try {
                        privateKeyBytes = await crypto.decryptChunk(decKey, privateKeyBytes);
                    } catch {
                        errorCallback();
                        return;
                    }

                    crypto.ingestProtectedKey(privateKeyBytes, privateKey => {
                        let publicKey = publicKeyRequest.result.key;
                        crypto.ingestPublicKey(publicKey, async publicKey => {
                            callback(privateKey, publicKey);
                        });
                    });
                }

                transaction.onerror = (event) => {
                    console.error("Error retrieving vault keys from IndexedDB:", event.target.error);
                    alert("Error fetching vault keys");
                    callback(null, null);
                };

                transaction.oncomplete = () => {
                    db.close();
                };
            }

            request.onerror = (event) => {
                console.error("Error opening local db");
                callback(null);
            }
        }

        /**
         * removeKeys removes all keys from the database and invokes the callback
         * with a boolean indicating if the removal was successful
         * @param callback {function(boolean)}
         */
        this.removeKeys = (callback) => {
            let request = indexedDB.open(this.dbName, this.dbVersion);

            request.onsuccess = (event) => {
                let db = event.target.result;
                let transaction = db.transaction([this.keysObjectStore], "readwrite");
                let objectStore = transaction.objectStore(this.keysObjectStore);

                let clearRequest = objectStore.clear();
                clearRequest.onsuccess = () => {
                    callback(true);
                }

                clearRequest.onerror = (event) => {
                    console.error("Error removing keys from IndexedDB:", event.target.error);
                    alert("Error removing keys from IndexedDB");
                    callback(false);
                };

                transaction.oncomplete = () => {
                    db.close();
                };
            }
        }
    }
}