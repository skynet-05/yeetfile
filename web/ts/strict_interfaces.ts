import * as crypto from "./crypto.js";
import {VaultUpload} from "./interfaces.js";

export class PackagedPassEntry {
    name: string;
    key: CryptoKey;
    upload: VaultUpload;
    entry: PassEntry;
    encData: Uint8Array;

    constructor(
        name: string,
        key: CryptoKey,
        upload: VaultUpload,
        entry: PassEntry,
        encData: Uint8Array,
    ) {
        this.name = name;
        this.key = key;
        this.upload = upload;
        this.entry = entry;
        this.encData = encData;
    }
}

export class PassEntry {
    username: string;
    password: string;
    passwordHistory: string[];
    urls: string[];
    notes: string;

    constructor(source: any = {}) {
        this.#init(source);
    }

    #init = (source: any = {}) => {
        if ("string" === typeof source) {
            try {
                source = JSON.parse(source);
            } catch (e) {
                throw new Error("Invalid JSON input: " + e);
            }
        }

        this.username = typeof source.username === "string" ? source.username : "";
        this.password = typeof source.password === "string" ? source.password : "";
        this.passwordHistory = Array.isArray(source.passwordHistory) &&
        source.passwordHistory.every(pw => typeof pw === "string") ?
            source.passwordHistory :
            [];
        this.urls = Array.isArray(source.urls) &&
        source.urls.every(url => typeof url === "string") ?
            source.urls :
            [];
        this.notes = typeof source.notes === "string" ? source.notes : "";
    }

    pack = async (
        name: string,
        folderID: string,
        encKey: Uint8Array,
        importedKey: CryptoKey,
    ): Promise<PackagedPassEntry> => {
        let str = JSON.stringify(this);
        let encoder = new TextEncoder();
        let encStr = await crypto.encryptChunk(importedKey, encoder.encode(str));
        let encName = await crypto.encryptChunk(importedKey, encoder.encode(name));
        let hexName = toHexString(encName);

        let upload = new VaultUpload();
        upload.name = hexName;
        upload.chunks = 1;
        upload.length = 1;
        upload.folderID = folderID;
        upload.protectedKey = Array.from(encKey);
        upload.passwordData = Array.from(encStr);

        return new PackagedPassEntry(name, importedKey, upload, this, encStr);
    }

    unpack = async (key: CryptoKey, data: Uint8Array) => {
        let decoder = new TextDecoder();
        let decData = await crypto.decryptChunk(key, data);
        let decoded = decoder.decode(decData);
        this.#init(decoded);
    }
}