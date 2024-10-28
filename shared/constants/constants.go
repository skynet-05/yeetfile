package constants

const VERSION = "1.0.0"

// JSSessionKey determines the hardcoded string that gets replaced in db.ts in
// order to have a semi-random value used in encrypting the user's key pair in
// IndexedDB.
// --- Changes to this value must be reflected in db.ts
const JSSessionKey = "JS_SESSION_KEY"

const CLIUserAgent = "yeetfile-cli"

const AuthSessionStore = "auth"

const Argon2Mem uint32 = 64 // MB
const Argon2Iter uint32 = 2

const LimiterSeconds = 30
const LimiterAttempts = 6

const TotalBandwidthMultiplier = 3 // 3x available storage
const BandwidthMonitorDuration = 7 // 7 day period

const IVSize int = 12
const KeySize int = 32
const ChunkSize int = 10000000 // 10 mb
const TotalOverhead int = 28   // encryption overhead (16) + iv size (12)
const MaxPlaintextLen = 2000
const MaxHintLen = 200
const PlaintextIDPrefix = "text"
const FileIDPrefix = "file"
const VerificationCodeLength = 6
const ChangeIDLength = 9
const MaxTransferThreads = 3
const MaxSendAgeDays = 30 //days
const MaxPassNoteLen = 500
const RecoveryCodeLen = 8
const SubMethodStripe = "stripe"
const SubMethodBTCPay = "btcpay"
