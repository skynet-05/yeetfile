const assert = require('node:assert').strict;
const crypto = require("../crypto.js");

const testDeriveSendingKey = testCallback => {
    // Test that two keys generated from the same password aren't the same
    crypto.deriveSendingKey("password", undefined, "", undefined, (firstKey, firstSalt) => {
        crypto.deriveSendingKey("password", undefined, "", undefined, async (secondKey, secondSalt) => {
           assert(firstSalt !== secondSalt);

           let firstKeyBytes = await crypto.exportKey(firstKey);
           let secondKeyBytes = await crypto.exportKey(secondKey);

           assert(firstKeyBytes !== secondKeyBytes);

           testCallback();
       });
    });
}

const testEncryptChunk = testCallback => {
    // Test data encryption
    let data = crypto.webcrypto.getRandomValues(new Uint8Array(10));
    crypto.deriveSendingKey("password", undefined, "", undefined, async key => {
        let blob = await crypto.encryptChunk(key, data);
        assert(blob !== data);
        assert(blob.length > data.length);

        // Encrypt the blob again, result should be different
        let blobNew = await crypto.encryptChunk(key, data);
        assert(blob !== blobNew);
        assert(blob.length === blobNew.length);

        testCallback();
    });
}

const testDecryptChunk = testCallback => {
    // Test data decryption
    let data = crypto.webcrypto.getRandomValues(new Uint8Array(10));
    crypto.deriveSendingKey("password", undefined, "", undefined, async key => {
        let encryptedBlob = await crypto.encryptChunk(key, data);
        let decryptedBlob = await crypto.decryptChunk(key, encryptedBlob);
        decryptedBlob = new Uint8Array(decryptedBlob);

        assert(data.length === decryptedBlob.length);
        for (let i in decryptedBlob) {
            assert(data[i] === decryptedBlob[i]);
        }

        crypto.deriveSendingKey("password", undefined, "", undefined, async newKey => {
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
        });
    });
}

const runTest = (testIdx) => {
    if (!testIdx) {
        testIdx = 0;
    }

    let testCallback = () => { runTest(testIdx + 1); }
    let tests = [
        testDeriveSendingKey,
        testEncryptChunk,
        testDecryptChunk,
    ]

    if (tests[testIdx]) {
        console.log("TEST: " + tests[testIdx].name);
        tests[testIdx](testCallback);
    }
}

runTest(0);