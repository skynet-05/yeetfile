import * as interfaces from "./interfaces.js";

type FolderCache = {
    [key: string]: interfaces.VaultFolderResponse;
};

export class VaultFolderCache {
    cache: FolderCache = {}

    constructor() {}

    get = (folderID: string): interfaces.VaultFolderResponse => {
        return this.cache[folderID]
    }

    set = (folderID: string, response: interfaces.VaultFolderResponse) => {
        this.cache[folderID] = response
    }

    addFolder = (folderID: string, folder: interfaces.VaultFolder) => {
        this.cache[folderID].folders.unshift(folder);
    }

    addItem = (folderID: string, item: interfaces.VaultItem) => {
        this.cache[folderID].items.unshift(item);
    }

    updateFolders = (folderID: string, folders: interfaces.VaultFolder[]) => {
        this.cache[folderID].folders = folders;
    }

    updateItems = (folderID: string, items: interfaces.VaultItem[]) => {
        this.cache[folderID].items = items;
    }
}