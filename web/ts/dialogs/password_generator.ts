import {generateRandomString, generatePassphrase} from "../crypto.js";
import {closeDialog} from "./dialogs.js";
import {YeetFileDB} from "../db.js";

export class PasswordGeneratorDialog {
    dialog: HTMLDialogElement;
    initState: string;

    passwordDisplay: HTMLElement;

    passwordLength: HTMLInputElement;
    passwordCapitalAZ: HTMLInputElement;
    passwordLowercaseAZ: HTMLInputElement;
    passwordNumbers: HTMLInputElement;
    passwordSymbols: HTMLInputElement;
    passwordSymbolsString: HTMLInputElement;

    passphraseWords: HTMLInputElement;
    passphraseShorterWords: HTMLInputElement;
    passphraseSeparator: HTMLInputElement;
    passphraseCapitalize: HTMLInputElement;
    passphraseNumber: HTMLInputElement;

    password: string;

    longWordlist: Array<string>;
    shortWordlist: Array<string>;

    constructor() {
        this.dialog = document.getElementById(
            "password-generator-dialog") as HTMLDialogElement;
        this.initState = this.dialog.innerHTML;

        let db = new YeetFileDB();
        db.fetchWordlists((success, long, short) => {
            if (!success || !long || !short) {
                alert("Wordlists not found -- passphrase generation may not work as expected");
                return;
            }

            this.longWordlist = long;
            this.shortWordlist = short;
        });
    }

    #regeneratePassword = () => {
        let pwLen = this.passwordLength.valueAsNumber;
        if (this.passwordLength.value.length === 0 || pwLen < 5 || pwLen > 128) {
            if (pwLen < 5) {
                this.passwordLength.valueAsNumber = 5;
            } else {
                this.passwordLength.valueAsNumber = 128;
            }
        }

        if (!this.passwordCapitalAZ.checked &&
            !this.passwordLowercaseAZ.checked &&
            !this.passwordNumbers.checked &&
            !this.passwordSymbols.checked
        ) {
            this.passwordLowercaseAZ.checked = true;
        }

        let generatedPassword = generateRandomString(
            this.passwordLength.valueAsNumber,
            this.passwordCapitalAZ.checked,
            this.passwordLowercaseAZ.checked,
            this.passwordNumbers.checked,
            this.passwordSymbols.checked,
            this.passwordSymbolsString.value);

        this.#updatePassword(generatedPassword);
    }

    #regeneratePassphrase = () => {
        let wordlist: Array<string>;
        if (this.passphraseShorterWords.checked) {
            wordlist = this.shortWordlist;
        } else {
            wordlist = this.longWordlist;
        }

        let passphraseWords = this.passphraseWords.valueAsNumber;
        if (passphraseWords < 2) {
            this.passphraseWords.valueAsNumber = 2;
        } else if (passphraseWords > 40) {
            this.passphraseWords.valueAsNumber = 40;
        }

        let passphraseSep = this.passphraseSeparator.value;
        if (passphraseSep.length > 15) {
            this.passphraseSeparator.value = this.passphraseSeparator.value.substring(0, 15);
        }

        let passphrase = generatePassphrase(
            wordlist,
            this.passphraseWords.valueAsNumber,
            this.passphraseSeparator.value,
            this.passphraseCapitalize.checked,
            this.passphraseNumber.checked);

        this.#updatePassword(passphrase);
    }

    #updatePassword = (password: string) => {
        this.password = password;
        this.passwordDisplay.innerText = password;

        let lengthCounter = document.getElementById("password-len");
        lengthCounter.innerText = `${password.length} characters`;
    }

    #init = () => {
        this.dialog.innerHTML = this.initState;

        this.passwordDisplay = document.getElementById("generated-password");

        this.passwordLength = document.getElementById("password-length") as HTMLInputElement;
        this.passwordCapitalAZ = document.getElementById("password-capital-az") as HTMLInputElement;
        this.passwordLowercaseAZ = document.getElementById("password-lowercase-az") as HTMLInputElement;
        this.passwordNumbers = document.getElementById("password-numbers") as HTMLInputElement;
        this.passwordSymbols = document.getElementById("password-symbols") as HTMLInputElement;
        this.passwordSymbolsString = document.getElementById("password-symbols-string") as HTMLInputElement;

        this.passphraseWords = document.getElementById("passphrase-words") as HTMLInputElement;
        this.passphraseShorterWords = document.getElementById("passphrase-shorter-words") as HTMLInputElement;
        this.passphraseSeparator = document.getElementById("passphrase-separator") as HTMLInputElement;
        this.passphraseCapitalize = document.getElementById("passphrase-capitalize") as HTMLInputElement;
        this.passphraseNumber = document.getElementById("passphrase-number") as HTMLInputElement;

        let passwordTable = document.getElementById("password-table") as HTMLTableElement;
        let passphraseTable = document.getElementById("passphrase-table") as HTMLTableElement;

        let passwordType = document.getElementById("password-type") as HTMLInputElement;
        let passphraseType = document.getElementById("passphrase-type") as HTMLInputElement;

        const setPasswordType = () => {
            this.passwordDisplay.className = "break-words";
            passphraseType.checked = false;
            passwordTable.style.display = "table";
            passphraseTable.style.display = "none";
            this.#regeneratePassword();
        }

        const setPassphraseType = () => {
            this.passwordDisplay.className = "";
            passwordType.checked = false;
            passwordTable.style.display = "none";
            passphraseTable.style.display = "table";
            this.#regeneratePassphrase();
        }

        passwordType.addEventListener("click", () => {
            setPasswordType();
        });

        passphraseType.addEventListener("click", () => {
            setPassphraseType();
        });

        setPasswordType();

        let generateBtn = document.getElementById("regenerate-password");
        generateBtn.addEventListener("click", () => {
            if (passwordType.checked) {
                this.#regeneratePassword();
            } else {
                this.#regeneratePassphrase();
            }
        });

        let cancelBtn = document.getElementById("cancel-generator");
        cancelBtn.addEventListener("click", () => {
            closeDialog(this.dialog);
        });
    }

    show = (callback?: (string) => void, cancelCallback?: () => void) => {
        this.#init();

        let submitBtn = document.getElementById("confirm-password");
        submitBtn.addEventListener("click", () => {
            closeDialog(this.dialog);
            callback(this.password);
        }, {once: true});

        if (!callback) {
            submitBtn.style.display = "none";
        } else {
            submitBtn.style.display = "flex";
        }

        let cancelBtn = document.getElementById("cancel-generator");
        cancelBtn.addEventListener("click", () => {
            closeDialog(this.dialog);
            if (cancelCallback) {
                cancelCallback();
            }
        });

        this.dialog.showModal();
    }
}