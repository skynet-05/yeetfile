package shared

const NonceSize int = 24
const KeySize int = 32
const ChunkSize int = 5242880 // 5 mb
const TotalOverhead int = 40  // secretbox overhead (16) + nonce size (24)
