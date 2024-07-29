package constants

const VERSION = "1.0.0"

const JSRandomSessionKey = "YEETFILE_RANDOM_SESSION_KEY"

const IVSize int = 12
const KeySize int = 32
const ChunkSize int = 10000000 // 10 mb
const TotalOverhead int = 28   // encryption overhead (16) + iv size (12)
const MaxPlaintextLen = 2000
const PlaintextIDPrefix = "text"
const FileIDPrefix = "file"
const VerificationCodeLength = 6
const MaxTransferThreads = 3
const MaxSendAgeDays = 30 //days
