// Auto-generated from shared/js.go. Don't edit this manually.

export const format = (endpoint, ...args) => {
    for (let arg of args) {
        endpoint = endpoint.replace("*", arg);
    }

    return endpoint;
}

export const VaultFile = "/api/v1/vault/file/*";
export const ShareFolder = "/api/v1/share/folder/*";
export const VerifyEmail = "/verify-email";
export const VaultFolder = "/api/v1/vault/folder/*";
export const UploadVaultFileMetadata = "/api/v1/vault/u";
export const UploadVaultFileData = "/api/v1/vault/u/*/*";
export const ShareFile = "/api/v1/share/file/*";
export const Signup = "/api/v1/signup";
export const Logout = "/api/v1/logout";
export const Session = "/api/v1/session";
export const VerifyAccount = "/api/v1/verify";
export const DownloadVaultFileMetadata = "/api/v1/vault/d/*";
export const Login = "/api/v1/login";
export const Forgot = "/api/v1/forgot";
export const Reset = "/api/v1/reset";
export const VaultRoot = "/api/v1/vault";
export const DownloadVaultFileData = "/api/v1/vault/d/*/*";
export const PubKey = "/api/v1/pubkey";
