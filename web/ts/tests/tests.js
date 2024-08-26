global.atob = function(base64) {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=';
  let str = '';
  let i = 0;
  let enc1, enc2, enc3, enc4;

  base64 = base64.replace(/[^A-Za-z0-9+/=]/g, '');

  while (i < base64.length) {
    enc1 = chars.indexOf(base64.charAt(i++));
    enc2 = chars.indexOf(base64.charAt(i++));
    enc3 = chars.indexOf(base64.charAt(i++));
    enc4 = chars.indexOf(base64.charAt(i++));

    str += String.fromCharCode((enc1 << 2) | (enc2 >> 4));

    if (enc3 !== 64) {
      str += String.fromCharCode(((enc2 & 15) << 4) | (enc3 >> 2));
    }

    if (enc4 !== 64) {
      str += String.fromCharCode(((enc3 & 3) << 6) | enc4);
    }
  }

  return str;
};

import { strict as assert } from "node:assert";
import * as crypto from "../crypto.js";

import { readFile } from 'fs/promises';
import path from 'path';
import vm from 'vm';
import { fileURLToPath } from 'url';

// Resolve __dirname in ES modules
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const sandbox = {
  window: {},
  document: {},
  crypto: global.crypto || require('crypto').webcrypto, // Use Node.js's webcrypto if available
};

// Simulate window and document objects
global.window = {};
global.document = {};

// Load the argon2-bundled.js script
const scriptPath = path.resolve(__dirname, '../../../node_modules/argon2-browser/dist/argon2-bundled.min.js');
const scriptContent = await readFile(scriptPath, 'utf-8');
// Use vm to run the script in the sandbox
vm.createContext(sandbox); // Create a context for the sandbox
vm.runInContext(scriptContent, sandbox); // Execute the script

// Extract the argon2 object from the sandbox
export const argon2 = sandbox.argon2;
console.log(argon2);

// Access argon2 from the global window object
//const { argon2 } = window;

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
        setTimeout(async () => {
            await tests[testIdx](testCallback);
        }, 500);
    }
}

runTest(0);
