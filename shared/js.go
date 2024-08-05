package shared

import (
	"fmt"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

const DBFilename = "db.js"

const constsJS = `
// Auto-generated from shared/js.go. Don't edit this manually.

export const IVSize = %d;
export const KeySize = %d;
export const ChunkSize = %d;
export const TotalOverhead = %d;
export const MaxPlaintextLen = %d;
export const PlaintextIDPrefix = "%s";
export const FileIDPrefix = "%s";
export const VerificationCodeLength = %d;`

const endpointsHeadJS = `
// Auto-generated from shared/js.go. Don't edit this manually.

export type Endpoint = {
    path: string
}

export class Endpoints {`

const endpointsTailJS = `

    static format(endpoint: Endpoint, ...args: string[]): string {
        let path = endpoint.path;
        for (let arg of args) {
            path = path.replace("*", arg);
        }

        return path;
    }
}
`

const endpointEntry = `
    static %s: Endpoint = {path: "%s"};`

func GenerateSharedJS() (string, string) {
	jsConsts := fmt.Sprintf(constsJS,
		constants.IVSize,
		constants.KeySize,
		constants.ChunkSize,
		constants.TotalOverhead,
		constants.MaxPlaintextLen,
		constants.PlaintextIDPrefix,
		constants.FileIDPrefix,
		constants.VerificationCodeLength)

	jsEndpoints := endpointsHeadJS
	for apiEndpoint, varName := range endpoints.JSVarNameMap {
		jsEndpoints += fmt.Sprintf(endpointEntry, varName, apiEndpoint)
	}

	jsEndpoints += endpointsTailJS
	return jsConsts, jsEndpoints
}
