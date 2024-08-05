import { strict as assert } from "node:assert";
import * as crypto from "../crypto.js";

const testDeriveSendingKey = async testCallback => {
    // Test that two keys generated from the same password aren't the same
    let [firstKey, firstSalt] = await crypto.deriveSendingKey("password", undefined, "");
    let [secondKey, secondSalt] = await crypto.deriveSendingKey("password", undefined, "");

    assert(firstSalt !== secondSalt);

    let firstKeyBytes = await crypto.exportKey(firstKey);
    let secondKeyBytes = await crypto.exportKey(secondKey);

    assert(firstKeyBytes !== secondKeyBytes);

    testCallback();
}

const testEncryptChunk = async testCallback => {
    // Test data encryption
    let data = crypto.webcrypto.getRandomValues(new Uint8Array(10));
    let [key, _] = await crypto.deriveSendingKey("password", undefined, "");

    let blob = await crypto.encryptChunk(key, data);
    assert(blob !== data);
    assert(blob.length > data.length);

    // Encrypt the blob again, result should be different
    let blobNew = await crypto.encryptChunk(key, data);
    assert(blob !== blobNew);
    assert(blob.length === blobNew.length);

    testCallback();
}

const testDecryptChunk = async testCallback => {
    // Test data decryption
    let data = crypto.webcrypto.getRandomValues(new Uint8Array(10));
    let [key, _salt] =  await crypto.deriveSendingKey("password", undefined, "");

    let encryptedBlob = await crypto.encryptChunk(key, data);
    let decryptedBlob = await crypto.decryptChunk(key, encryptedBlob);
    decryptedBlob = new Uint8Array(decryptedBlob);

    assert(data.length === decryptedBlob.length);
    for (let i in decryptedBlob) {
        assert(data[i] === decryptedBlob[i]);
    }

    let [newKey, _newSalt] = await crypto.deriveSendingKey("password", undefined, "");

    // Attempt decrypting with new key
    crypto.decryptChunk(newKey, encryptedBlob).then(() => {
        // This indicates that a chunk encrypted with one key was decrypted with another key,
        // which would always indicate failure
        assert(false);
    }).catch(err => {
        // This indicates that a chunk encrypted with one key could not be decrypted by a different
        // key, which indicates success
        assert(err);
        testCallback();
    });
}

const testLoginKeyHash = async testCallback => {
    let userKey = await crypto.generateUserKey("myemail@domain.com", "mypassword");
    let loginKeyHash = await crypto.generateLoginKeyHash(userKey, "mypassword");
    let loginKeyHashDuplicate = await crypto.generateLoginKeyHash(userKey, "mypassword");

    assert(loginKeyHash.length === loginKeyHashDuplicate.length)
    for (let i in loginKeyHash) {
        assert(loginKeyHash[i] === loginKeyHashDuplicate[i]);
    }

    testCallback();
}

const testKeyPair = async testCallback => {
    let userKey = await crypto.generateUserKey("myemail@domain.com", "mypassword");
    let keyPair = await crypto.generateKeyPair();

    let publicKey = await crypto.exportKey(keyPair.publicKey, "spki");
    let privateKey = await crypto.exportKey(keyPair.privateKey, "pkcs8");
    let protectedKey = await crypto.encryptChunk(userKey, privateKey);
    assert(publicKey);
    assert(privateKey);
    assert(protectedKey);

    let folderKey = await crypto.generateRandomKey();
    let protectedRootFolderKey = await crypto.encryptRSA(keyPair.publicKey, folderKey);

    let rootFolderKey = await crypto.decryptRSA(keyPair.privateKey, protectedRootFolderKey);
    let folderKeyImport = await crypto.importKey(rootFolderKey);
    assert(folderKeyImport);

    testCallback();
}

const runTest = async (testIdx) => {
    if (!testIdx) {
        testIdx = 0;
    }

    let testCallback = () => { runTest(testIdx + 1); }
    let tests = [
        testDeriveSendingKey,
        testEncryptChunk,
        testDecryptChunk,
        testLoginKeyHash,
        testKeyPair,
    ]

    if (tests[testIdx]) {
        console.log("TEST: " + tests[testIdx].name);
        await tests[testIdx](testCallback);
    }
}

runTest(0);
