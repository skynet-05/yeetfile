class YeetFileDB {
    constructor() {
        this.dbName = "yeetfileDB";
        this.dbVersion = 1;
        this.keysObjectStore = "keys";
        this.privateKeyID = 1;
        this.publicKeyID = 2;

        const request = indexedDB.open(this.dbName, this.dbVersion);

        request.onupgradeneeded = (event) => {
            const db = event.target.result;
            db.createObjectStore(this.keysObjectStore, { keyPath: "id" });
        };

        /**
         * insertKey inserts a non-exportable CryptoKey object into the
         * indexeddb database
         * @param privateKey {CryptoKey}
         * @param publicKey {CryptoKey}
         * @param callback {function(boolean)}
         */
        this.insertVaultKeyPair = (privateKey, publicKey) => {
            this.removeKeys(() => {});

            let request = indexedDB.open(this.dbName, this.dbVersion);
            request.onsuccess = (event) => {
                let db = event.target.result;

                let transaction = db.transaction([this.keysObjectStore], "readwrite");
                let objectStore = transaction.objectStore(this.keysObjectStore);

                objectStore.put({ id: this.privateKeyID, key: privateKey });
                objectStore.put({ id: this.publicKeyID, key: publicKey });

                transaction.onerror = (event) => {
                    console.error("Error adding vault keys to IndexedDB:", event.target.error);
                    alert("Error preparing vault keys");
                };

                transaction.oncomplete = () => {
                    db.close();
                };
            }

            request.onerror = () => {
                console.error("Error opening local db");
            }
        }

        /**
         * getVaultKey returns the vault key from the indexeddb, if it's available
         * @param callback {function(CryptoKey, CryptoKey)}
         */
        this.getVaultKeyPair = (callback) => {
            let request = indexedDB.open(this.dbName, this.dbVersion);

            request.onsuccess = (event) => {
                let db = event.target.result;

                let transaction = db.transaction([this.keysObjectStore], "readonly");
                let objectStore = transaction.objectStore(this.keysObjectStore);

                let privateKeyRequest = objectStore.get(this.privateKeyID);

                privateKeyRequest.onsuccess = (event) => {
                    let privateKey = event.target.result.key;

                    let publicKeyRequest = objectStore.get(this.publicKeyID);
                    publicKeyRequest.onsuccess = (event) => {
                        let publicKey = event.target.result.key;

                        callback(privateKey, publicKey);
                    }
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