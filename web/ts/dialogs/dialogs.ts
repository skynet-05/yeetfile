export enum DialogSignal {
    Cancel = 0,
    Delete,
    Download,
    Remove,
    Rename,
    Share,
}

/**
 * Close all dialogs
 */
export const closeDialogs = (): void => {
    let dialogs = document.getElementsByTagName("dialog");
    for (let i = 0; i < dialogs.length; i++) {
        closeDialog(dialogs[i]);
    }
}

/**
 * Close specific HTMLDialogElement
 * @param dialog
 */
export const closeDialog = (dialog: HTMLDialogElement): void => {
    dialog.close();

    // Reset listeners for dynamic dialogs
    if (dialog.dataset.dynamic) {
        (dialog as HTMLDialogElement).outerHTML = dialog.outerHTML;
    }
}

/**
 * Determine if a dialog is currently open on the page.
 */
export const isDialogOpen = (): boolean => {
    let dialogs = document.getElementsByTagName("dialog");
    for (let i = 0; i < dialogs.length; i++) {
        if (!dialogs[i].open) {
            continue;
        }

        return true;
    }

    return false;
}

/**
 * Enter key submits open dialog, Esc key closes open dialog
 */
document.addEventListener("keydown", (event: KeyboardEvent) => {
    if (event.key === "Enter" || event.key === "Escape") {
        if (document.activeElement.tagName === "TEXTAREA") {
            return;
        }

        let dialogs = document.getElementsByTagName("dialog");
        for (let i = 0; i < dialogs.length; i++) {
            if (!dialogs[i].open) {
                continue;
            }

            let dialog = dialogs[i];
            let buttons = dialog.getElementsByTagName("button");
            for (let j = 0; j < buttons.length; j++) {
                if ((event.key === "Enter" && buttons[j].id.startsWith("submit")) ||
                    (event.key === "Escape" && buttons[j].id.startsWith("cancel"))) {
                    event.preventDefault();
                    buttons[j].click();
                    return;
                }
            }
        }
    }
});