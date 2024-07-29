package shared

import (
	"fmt"
	"yeetfile/shared/constants"
	"yeetfile/shared/endpoints"
)

const ConstsFilename = "constants.js"
const EndpointsFilename = "endpoints.js"
const DBFilename = "db.js"

const constsJS = `// Auto-generated from shared/js.go. Don't edit this manually.

export const IVSize = %d;
export const KeySize = %d;
export const ChunkSize = %d;
export const TotalOverhead = %d;
export const MaxPlaintextLen = %d;
export const PlaintextIDPrefix = "%s";
export const FileIDPrefix = "%s";
export const VerificationCodeLength = %d;`

const endpointsJS = `// Auto-generated from shared/js.go. Don't edit this manually.

export const format = (endpoint, ...args) => {
    for (let arg of args) {
        endpoint = endpoint.replace("*", arg);
    }

    return endpoint;
}

`

const endpointStr = `export const %s = "%s";
`

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

	jsEndpoints := endpointsJS
	for apiEndpoint, varName := range endpoints.JSVarNameMap {
		jsEndpoints += fmt.Sprintf(endpointStr, varName, apiEndpoint)
	}

	return jsConsts, jsEndpoints
}
