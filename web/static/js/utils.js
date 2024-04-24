const chunkSize = 10000000;

const expUnits = {
    minutes: 0,
    hours: 1,
    days: 2
}

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

const calcFileSize = bytes => {
    let thresh = 1000;

    if (Math.abs(bytes) < thresh) {
        return bytes + ' B';
    }

    const units = ['KB', 'MB', 'GB', 'TB'];
    let u = -1;
    const r = 10;

    do {
        bytes /= thresh;
        ++u;
    } while (Math.round(Math.abs(bytes) * r) / r >= thresh && u < units.length - 1);


    return bytes.toFixed(1) + ' ' + units[u];
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

const arrayToBase64 = async buffer => {
    const base64url = await new Promise(r => {
        const reader = new FileReader()
        reader.onload = () => r(reader.result)
        reader.readAsDataURL(new Blob([buffer]))
    });
    return base64url.slice(base64url.indexOf(',') + 1);
}

const formatDate = date => {
    let localDate = new Date(date);
    return localDate.toLocaleString();
}

const validateExpiration = (exp, unit) => {
    let maxDays = 30;
    let maxHours = 24 * maxDays;
    let maxMinutes = 60 * maxHours;

    if (unit === expUnits.minutes) {
        if (exp <= 0 || exp > maxMinutes) {
            alert(`Expiration must be between 0-${maxMinutes} minutes`);
            return false;
        }
    }

    if (unit === expUnits.hours) {
        if (exp <= 0 || exp > maxHours) {
            alert(`Expiration must be between 0-${maxHours} hours`);
            return false;
        }
    }

    if (unit === expUnits.days) {
        if (exp <= 0 || exp > maxDays) {
            alert(`Expiration must be between 0-${maxDays} days`);
            return false;
        }
    }

    return true;
}

const updateButton = (btn, disabled, label) => {
    btn.disabled = disabled;
    btn.innerText = label;
}