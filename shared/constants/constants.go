package constants

const VERSION = "1.0.0"

const JSRandomSessionKey = "YEETFILE_RANDOM_SESSION_KEY"

const CLIUserAgent = "yeetfile-cli"

const AuthSessionStore = "auth"

const Argon2Mem uint32 = 64 * 1024 // 64MB
const Argon2Iter uint32 = 2

const LimiterSeconds = 30
const LimiterAttempts = 3

const IVSize int = 12
const KeySize int = 32
const ChunkSize int = 10000000 // 10 mb
const TotalOverhead int = 28   // encryption overhead (16) + iv size (12)
const MaxPlaintextLen = 2000
const MaxHintLen = 200
const PlaintextIDPrefix = "text"
const FileIDPrefix = "file"
const VerificationCodeLength = 6
const MaxTransferThreads = 3
const MaxSendAgeDays = 30 //days
const SubMethodStripe = "stripe"
const SubMethodBTCPay = "btcpay"
