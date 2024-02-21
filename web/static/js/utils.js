const chunkSize = 5242880;

const hexToBytes = (hexString) =>
    Uint8Array.from(hexString.match(/.{1,2}/g).map((byte) => parseInt(byte, 16)));

const toHexString = (bytes) =>
    bytes.reduce((str, byte) => str + byte.toString(16).padStart(2, '0'), '');

const getNumChunks = (size) =>
    Math.ceil(size / chunkSize);

const genRandomString = (size) =>
    (Math.random().toString(36)+'00000000000000000').slice(2, size+2)

const concatTypedArrays = (a, b) => {
    let c = new (a.constructor)(a.length + b.length);
    c.set(a, 0);
    c.set(b, a.length);
    return c;
}

const getExpString = (exp, units) => {
    switch (units) {
        case expUnits.minutes:
            return `${exp}m`
        case expUnits.hours:
            return `${exp}h`
        case expUnits.days:
            return `${exp}d`
    }
}

const base64ToArray = (base64) => {
    let binaryString = atob(base64);
    let bytes = new Uint8Array(binaryString.length);
    for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i);
    }
    return bytes;
}