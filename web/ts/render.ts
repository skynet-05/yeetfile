const nonTextFileTypes = [
    // Image types
    "png", "jpg", "jpeg", "svg",
    // Audio types
    "wav", "mp3",
    // Video types
    "mp4", "webm",
    // Other
    "pdf",
]

const getExt = (filename: string) => {
    let extSplit = filename.split(".");
    return extSplit[extSplit.length - 1].toLowerCase();
}

const getMimetypeFromExt = (ext: string): string => {
    switch (ext) {
        case "mp3":
            return "mpeg";
        case "svg":
            return "svg+xml"
        default:
            return ext;
    }
}

export const renderFileHTML = (filename: string, bytes: Uint8Array, callback: (string, MediaSource) => void) => {
    let ext = getExt(filename);
    let mime = getMimetypeFromExt(ext);
    let src = "REPLACE_SRC";

    let type;
    let tag;
    switch (ext) {
        case "jpeg":
        case "jpg":
        case "png":
        case "svg":
            type = {type: `image/${mime}`};
            tag = `<img src="${src}" alt="${filename}"/>`;
            break;
        case "mp3":
        case "wav":
            type = {type: `audio/${mime}`};
            tag = `<audio controls><source src="${src}" type="audio/${mime}">Unsupported format</audio>`;
            break;
        case "mp4":
        case "webm":
            type = {type: `video/${mime}`}
            tag = `<video controls><source src="${src}" type="video/${mime}">Unsupported format</video>`;
            break;
        case "pdf":
            type = {type :`application/${mime}`}
            tag = `<iframe src="${src}" type="application/${mime}">`;
            break;
    }

    let blob = new Blob([bytes], type);
    let blobURL = URL.createObjectURL(blob);
    tag = tag.replace(src, blobURL);
    callback(tag, blobURL);
}

export const isNonTextFileType = (filename: string) => {
    let ext = getExt(filename);
    return nonTextFileTypes.includes(ext);
}