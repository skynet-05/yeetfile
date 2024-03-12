package shared

const NonceSize int = 12
const KeySize int = 32
const ChunkSize int = 5242880 // 5 mb
const TotalOverhead int = 28  // encryption overhead (16) + nonce size (12)
const MaxPlaintextLen = 2000
const PlaintextIDPrefix = "text"
const FileIDPrefix = "file"
