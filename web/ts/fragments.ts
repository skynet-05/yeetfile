
export const VaultFileViewDiv = (): string => {
    return `
    <div id="vault-file-header"></div>
    <div id="vault-file-content"></div>
    <pre id="vault-text-wrapper"><code id="vault-text-content"></code></pre>`;
}

export const LoadingSpinner = (align: string): string => {
    return `<img class="vert-align-${align} small-icon progress-spinner" src="/static/icons/progress.svg">`
}