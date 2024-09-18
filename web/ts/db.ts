import * as crypto from "./crypto.js";
import {JS_SESSION_KEY} from "./constants.js";

export class YeetFileDB {
    private readonly dbName: string;
    private readonly dbVersion: number;
    private readonly keysObjectStore: string;
    private readonly privateKeyID: number;
    private readonly publicKeyID: number;
    private readonly passwordProtectedID: number;

    isPasswordProtected: (callback: (isPwProtected: boolean) => void) => void;
    insertVaultKeyPair: (
        privateKey: Uint8Array,
        publicKey: Uint8Array,
        password: string,
        callback: (success: boolean) => void,
    ) => void;
    getVaultKeyPair: (
        password: string,
        callback: (privKey: CryptoKey, pubKey: CryptoKey) => void,
        errorCallback: () => void,
    ) => void;
    removeKeys: (callback: (success: boolean) => void) => void;

    constructor() {
        this.dbName = "yeetfileDB";
        this.dbVersion = 1;
        this.keysObjectStore = "keys";
        this.privateKeyID = 1;
        this.publicKeyID = 2;
        this.passwordProtectedID = 3;

        const request = indexedDB.open(this.dbName, this.dbVersion);

        request.onupgradeneeded = (event: IDBVersionChangeEvent) => {
            const db = (event.target as IDBOpenDBRequest)?.result;
            if (db) {
                db.createObjectStore(this.keysObjectStore, { keyPath: "id" });
            }
        };

        /**
         * insertKey encrypts the user's private key with a random key and stores
         * the bytes for the private and public keys in indexeddb
         * @param privateKey {Uint8Array}
         * @param publicKey {Uint8Array}
         * @param password {string}
         * @param callback {function(boolean)}
         */
        this.insertVaultKeyPair = async (
            privateKey: Uint8Array,
            publicKey: Uint8Array,
            password: string,
            callback: (arg: boolean) => void,
        ) => {
            this.removeKeys(() => {});

            let encKey;
            if (password.length > 0) {
                encKey = await crypto.generateArgon2Key(password, JS_SESSION_KEY);
            } else {
                encKey = await crypto.importKey(hexToBytes(JS_SESSION_KEY))
            }

            // Replaced w/ random value on each request (needs to be cached by browser)
            let encPrivKey = await crypto.encryptChunk(encKey, privateKey);

            let request = indexedDB.open(this.dbName, this.dbVersion);
            request.onsuccess = async (event: Event) => {
                const db = (event.target as IDBOpenDBRequest)?.result;
                if (!db) {
                    callback(false);
                    return;
                }

                let transaction = db.transaction([this.keysObjectStore], "readwrite");
                let objectStore = transaction.objectStore(this.keysObjectStore);
                try {
                    let putPrivateKeyRequest = await objectStore.put({
                        id: this.privateKeyID,
                        key: encPrivKey
                    });
                    putPrivateKeyRequest.onerror = (event: Event) => {
                        const error = (event.target as IDBRequest).error;
                        console.error("Error storing private key:", error);
                        callback(false);
                    };

                    let putPublicKeyRequest = await objectStore.put({
                        id: this.publicKeyID,
                        key: publicKey
                    });
                    putPublicKeyRequest.onerror = (event: Event) => {
                        const error = (event.target as IDBRequest).error;
                        console.error("Error storing public key:", error);
                        callback(false);
                    };

                    let putPasswordProtectedRequest = await objectStore.put({
                        id: this.passwordProtectedID,
                        key: password.length > 0
                    });
                    putPasswordProtectedRequest.onerror = (event: Event) => {
                        const error = (event.target as IDBRequest).error;
                        console.error("Error storing pw protection flag:", error);
                        callback(false);
                    };
                } catch (error) {
                    console.error("Error during put operations:", error);
                }

                transaction.onerror = (event: Event) => {
                    const error = (event.target as IDBRequest).error;
                    console.error("Error adding vault keys to IndexedDB:", error);
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
                const db = (event.target as IDBOpenDBRequest)?.result;
                if (!db) {
                    callback(false);
                    return;
                }

                let transaction = db.transaction([this.keysObjectStore], "readonly");
                let objectStore = transaction.objectStore(this.keysObjectStore);
                let passwordProtectedRequest = objectStore.get(this.passwordProtectedID);

                passwordProtectedRequest.onsuccess = (event) => {
                    const result = (event.target as IDBRequest).result;
                    callback(result.key);
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
                const db = (event.target as IDBOpenDBRequest)?.result;
                if (!db) {
                    errorCallback();
                    return;
                }

                let decKey;
                if (password.length > 0) {
                    decKey = await crypto.generateArgon2Key(password, JS_SESSION_KEY);
                } else {
                    decKey = await crypto.importKey(hexToBytes(JS_SESSION_KEY));
                }

                let transaction = db.transaction([this.keysObjectStore], "readonly");
                let objectStore = transaction.objectStore(this.keysObjectStore);
                let privateKeyRequest = objectStore.get(this.privateKeyID);
                let publicKeyRequest = objectStore.get(this.publicKeyID);

                privateKeyRequest.onsuccess = async (event) => {
                    const result = (event.target as IDBRequest).result;
                    let privateKeyBytes = result.key;

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
                    const error = (event.target as IDBRequest).error;
                    console.error("Error retrieving vault keys from IndexedDB:",
                        error);
                    alert("Error fetching vault keys");
                    errorCallback();
                };

                transaction.oncomplete = () => {
                    db.close();
                };
            }

            request.onerror = () => {
                console.error("Error opening local db");
                errorCallback();
            }
        }

        /**
         * removeKeys removes all keys from the database and invokes the callback
         * with a boolean indicating if the removal was successful
         * @param callback {function(boolean)}
         */
        this.removeKeys = (callback) => {
            let request = indexedDB.open(this.dbName, this.dbVersion);

            request.onsuccess = (event: Event) => {
                const db = (event.target as IDBOpenDBRequest)?.result;
                if (!db) {
                    callback(false);
                    return;
                }

                let transaction = db.transaction([this.keysObjectStore], "readwrite");
                let objectStore = transaction.objectStore(this.keysObjectStore);

                let clearRequest = objectStore.clear();
                clearRequest.onsuccess = () => {
                    callback(true);
                }

                clearRequest.onerror = (event) => {
                    const error = (event.target as IDBRequest).error;
                    console.error("Error removing keys from IndexedDB:", error);
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