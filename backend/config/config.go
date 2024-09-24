package config

import (
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	"log"
	"os"
	"slices"
	"strings"
	"yeetfile/backend/utils"
	"yeetfile/shared"
	"yeetfile/shared/constants"
)

// =============================================================================
// General configuration
// =============================================================================

const LocalStorage = "local"
const B2Storage = "b2"

var defaultSecret = []byte("yeetfile-debug-secret-key-123456")
var storageType = utils.GetEnvVar("YEETFILE_STORAGE", LocalStorage)
var domain = os.Getenv("YEETFILE_DOMAIN")
var defaultUserStorage = utils.GetEnvVarInt("YEETFILE_DEFAULT_USER_STORAGE", -1)
var defaultUserSend = utils.GetEnvVarInt("YEETFILE_DEFAULT_USER_SEND", -1)
var maxNumUsers = utils.GetEnvVarInt("YEETFILE_MAX_NUM_USERS", -1)
var password = []byte(utils.GetEnvVar("YEETFILE_SERVER_PASSWORD", ""))
var secret = utils.GetEnvVarBytesB64("YEETFILE_SERVER_SECRET", defaultSecret)
var fallbackWebSecret = utils.GetEnvVarBytesB64(
	"YEETFILE_FALLBACK_WEB_SECRET",
	securecookie.GenerateRandomKey(32))

var IsDebugMode = utils.GetEnvVarBool("YEETFILE_DEBUG", false)

// =============================================================================
// Email configuration (used in account verification and billing reminders)
// =============================================================================

type EmailConfig struct {
	Configured bool
	Address    string
	Host       string
	Port       string
	Password   string
}

var email = EmailConfig{
	Configured: false,
	Address:    os.Getenv("YEETFILE_EMAIL_ADDR"),
	Host:       os.Getenv("YEETFILE_EMAIL_HOST"),
	Port:       os.Getenv("YEETFILE_EMAIL_PORT"),
	Password:   os.Getenv("YEETFILE_EMAIL_PASSWORD"),
}

// =============================================================================
// Billing configuration (Stripe)
// =============================================================================

type StripeBillingConfig struct {
	Configured    bool
	Key           string
	WebhookSecret string
	PortalLink    string

	SubNoviceMonthly     string
	SubNoviceMonthlyLink string
	SubNoviceYearly      string
	SubNoviceYearlyLink  string

	SubRegularMonthly     string
	SubRegularMonthlyLink string
	SubRegularYearly      string
	SubRegularYearlyLink  string

	SubAdvancedMonthly     string
	SubAdvancedMonthlyLink string
	SubAdvancedYearly      string
	SubAdvancedYearlyLink  string

	//Add50GBSend     string
	//Add50GBSendLink string
	//
	//Add100GBSend     string
	//Add100GBSendLink string
	//
	//Add250GBSend     string
	//Add250GBSendLink string
}

var stripeBilling = StripeBillingConfig{
	Key:           os.Getenv("YEETFILE_STRIPE_KEY"),
	WebhookSecret: os.Getenv("YEETFILE_STRIPE_WEBHOOK_SECRET"),
	PortalLink:    os.Getenv("YEETFILE_STRIPE_PORTAL_LINK"),

	SubNoviceMonthly:     os.Getenv("YEETFILE_STRIPE_SUB_NOVICE_MONTHLY"),
	SubNoviceMonthlyLink: os.Getenv("YEETFILE_STRIPE_SUB_NOVICE_MONTHLY_LINK"),
	SubNoviceYearly:      os.Getenv("YEETFILE_STRIPE_SUB_NOVICE_YEARLY"),
	SubNoviceYearlyLink:  os.Getenv("YEETFILE_STRIPE_SUB_NOVICE_YEARLY_LINK"),

	SubRegularMonthly:     os.Getenv("YEETFILE_STRIPE_SUB_REGULAR_MONTHLY"),
	SubRegularMonthlyLink: os.Getenv("YEETFILE_STRIPE_SUB_REGULAR_MONTHLY_LINK"),
	SubRegularYearly:      os.Getenv("YEETFILE_STRIPE_SUB_REGULAR_YEARLY"),
	SubRegularYearlyLink:  os.Getenv("YEETFILE_STRIPE_SUB_REGULAR_YEARLY_LINK"),

	SubAdvancedMonthly:     os.Getenv("YEETFILE_STRIPE_SUB_ADVANCED_MONTHLY"),
	SubAdvancedMonthlyLink: os.Getenv("YEETFILE_STRIPE_SUB_ADVANCED_MONTHLY_LINK"),
	SubAdvancedYearly:      os.Getenv("YEETFILE_STRIPE_SUB_ADVANCED_YEARLY"),
	SubAdvancedYearlyLink:  os.Getenv("YEETFILE_STRIPE_SUB_ADVANCED_YEARLY_LINK"),
}

// =============================================================================
// Billing configuration (BTCPay)
// =============================================================================

type BTCPayBillingConfig struct {
	Configured    bool
	APIKey        string
	WebhookSecret string
	StoreID       string
	ServerURL     string

	SubNoviceMonthlyLink   string
	SubNoviceYearlyLink    string
	SubRegularMonthlyLink  string
	SubRegularYearlyLink   string
	SubAdvancedMonthlyLink string
	SubAdvancedYearlyLink  string
}

var btcPayBilling = BTCPayBillingConfig{
	APIKey:        os.Getenv("YEETFILE_BTCPAY_API_KEY"),
	WebhookSecret: os.Getenv("YEETFILE_BTCPAY_WEBHOOK_SECRET"),
	StoreID:       os.Getenv("YEETFILE_BTCPAY_STORE_ID"),
	ServerURL:     os.Getenv("YEETFILE_BTCPAY_SERVER_URL"),

	SubNoviceMonthlyLink:   os.Getenv("YEETFILE_BTCPAY_SUB_NOVICE_MONTHLY_LINK"),
	SubNoviceYearlyLink:    os.Getenv("YEETFILE_BTCPAY_SUB_NOVICE_YEARLY_LINK"),
	SubRegularMonthlyLink:  os.Getenv("YEETFILE_BTCPAY_SUB_REGULAR_MONTHLY_LINK"),
	SubRegularYearlyLink:   os.Getenv("YEETFILE_BTCPAY_SUB_REGULAR_YEARLY_LINK"),
	SubAdvancedMonthlyLink: os.Getenv("YEETFILE_BTCPAY_SUB_ADVANCED_MONTHLY_LINK"),
	SubAdvancedYearlyLink:  os.Getenv("YEETFILE_BTCPAY_SUB_ADVANCED_YEARLY_LINK"),
}

// =============================================================================
// Full server config
// =============================================================================

type ServerConfig struct {
	StorageType        string
	Domain             string
	DefaultUserStorage int
	DefaultUserSend    int
	MaxUserCount       int
	CurrentUserCount   int
	Email              EmailConfig
	StripeBilling      StripeBillingConfig
	BTCPayBilling      BTCPayBillingConfig
	BillingEnabled     bool
	Version            string
	PasswordHash       []byte
	ServerSecret       []byte
	FallbackWebSecret  []byte
}

type TemplateConfig struct {
	Version          string
	CurrentUserCount int
	MaxUserCount     int
	EmailEnabled     bool
	BillingEnabled   bool
	StripeEnabled    bool
	BTCPayEnabled    bool
}

var YeetFileConfig ServerConfig
var HTMLConfig TemplateConfig

func init() {
	email.Configured = !utils.IsStructMissingAnyField(email)
	stripeBilling.Configured = email.Configured &&
		!utils.IsStructMissingAnyField(stripeBilling)
	btcPayBilling.Configured = email.Configured &&
		!utils.IsStructMissingAnyField(btcPayBilling)

	var passwordHash []byte
	var err error
	if len(password) > 0 {
		passwordHash, err = bcrypt.GenerateFromPassword(password, 8)
		if err != nil {
			panic(err)
		}
	}

	if slices.Equal(secret, defaultSecret) {
		logWarning(
			"Server secret is set to the default value.",
			"YEETFILE_SERVER_SECRET should be set to a ",
			"unique, 32-byte base-64 encoded value in production.")
	} else if len(secret) != constants.KeySize {
		log.Fatalf("ERROR: YEETFILE_SERVER_SECRET is %d bytes, but %d "+
			"bytes are required.", len(secret), constants.KeySize)
	}

	YeetFileConfig = ServerConfig{
		StorageType:        storageType,
		Domain:             domain,
		DefaultUserStorage: defaultUserStorage,
		DefaultUserSend:    defaultUserSend,
		MaxUserCount:       maxNumUsers,
		Email:              email,
		StripeBilling:      stripeBilling,
		BTCPayBilling:      btcPayBilling,
		BillingEnabled:     stripeBilling.Configured || btcPayBilling.Configured,
		Version:            constants.VERSION,
		PasswordHash:       passwordHash,
		ServerSecret:       secret,
		FallbackWebSecret:  fallbackWebSecret,
	}

	// Subset of main server config to use in HTML templating
	HTMLConfig = TemplateConfig{
		Version:        YeetFileConfig.Version,
		MaxUserCount:   YeetFileConfig.MaxUserCount,
		EmailEnabled:   YeetFileConfig.Email.Configured,
		BillingEnabled: YeetFileConfig.BillingEnabled,
		StripeEnabled:  YeetFileConfig.StripeBilling.Configured,
		BTCPayEnabled:  YeetFileConfig.BTCPayBilling.Configured,
	}

	log.Printf("Configuration:\n"+
		"  Email:            %v\n"+
		"  Billing (Stripe): %v\n"+
		"  Billing (BTCPay): %v\n",
		email.Configured,
		stripeBilling.Configured,
		btcPayBilling.Configured,
	)

	if IsDebugMode {
		logWarning(
			"DEBUG MODE IS ACTIVE!",
			"DO NOT USE THIS SETTING IN PRODUCTION!")
	}
}

func logWarning(warnings ...string) {
	log.Println(strings.Repeat("@", 57))
	for _, warning := range warnings {
		log.Printf("!!! " + warning + "\n")
	}
	log.Println(strings.Repeat("@", 57))
}

func GetServerInfoStruct() shared.ServerInfo {
	var storageBackend string
	if storageType == B2Storage {
		storageBackend = "Backblaze B2"
	} else {
		storageBackend = "Server Storage"
	}

	return shared.ServerInfo{
		StorageBackend:     storageBackend,
		PasswordRestricted: YeetFileConfig.PasswordHash != nil,
		MaxUserCountSet:    YeetFileConfig.MaxUserCount > 0,
		EmailConfigured:    YeetFileConfig.Email.Configured,
		BillingEnabled:     YeetFileConfig.BillingEnabled,
		StripeEnabled:      YeetFileConfig.BTCPayBilling.Configured,
		BTCPayEnabled:      YeetFileConfig.StripeBilling.Configured,
		DefaultStorage:     YeetFileConfig.DefaultUserStorage,
		DefaultSend:        YeetFileConfig.DefaultUserSend,
	}
}
