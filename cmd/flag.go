package cmd // import "breve.us/authsvc/cmd"

import "github.com/urfave/cli"

const (
	realm = "breve.us/authsvc"

	logPrefixAuth = "[authsrv] "

	listenPort  = "port"
	listenIP    = "bind"
	debug       = "debug"
	corsOrigins = "origins"
	insecure    = "insecure"
	publicHome  = "public"
	crypthash   = "hash"
	cryptblock  = "block"
	cacheDir    = "cache"
	loginPath   = "login"

	ldapHost      = "ldapHost"
	ldapPort      = "ldapPort"
	ldapTLS       = "ldapTLS"
	ldapAdminUser = "ldapAdminUser"
	ldapAdminPass = "ldapAdminPass"
	ldapBaseDN    = "ldapBaseDN"

	authRoot  = "/auth/"
	oauthRoot = "/oauth/"
)

var (
	portFlag = cli.IntFlag{
		Name:   listenPort,
		Usage:  "listen port",
		EnvVar: "PORT",
		Value:  42001,
	}
	bindFlag = cli.StringFlag{
		Name:   listenIP,
		Usage:  "IP address to bind listener to",
		EnvVar: "BIND_IP",
		Value:  "127.0.0.1",
	}
	debugFlag = cli.BoolFlag{
		Name:   debug,
		Usage:  "extra debug logging",
		EnvVar: "DEBUG",
	}
	corsOriginsFlag = cli.StringSliceFlag{
		Name:   corsOrigins,
		Usage:  "CORS Origins values",
		EnvVar: "CORS_ORIGINS",
		Value:  &cli.StringSlice{"*"},
	}
	insecureFlag = cli.BoolFlag{
		Name:   insecure,
		Usage:  "don't use secure cookies (Security Risk, but helpful when serving on localhost, or non-TLS endpoints)",
		EnvVar: "INSECURE",
	}
	publicHomeFlag = cli.StringFlag{
		Name:   publicHome,
		Usage:  "path to public folder",
		EnvVar: "PUBLIC_HOME",
		Value:  "public",
	}
	hashFlag = cli.StringFlag{
		Name:   crypthash,
		Usage:  "crypt hash string (for secure cookies)",
		EnvVar: "CRYPT_HASH",
	}
	blockFlag = cli.StringFlag{
		Name:   cryptblock,
		Usage:  "crypt block string (for secure cookies)",
		EnvVar: "CRYPT_BLOCK",
	}
	cacheDirFlag = cli.StringFlag{
		Name:   cacheDir,
		Usage:  "optional directory for persistent caches (if this is empty, or not a valid directory, in-memory caches will be used)",
		EnvVar: "CACHE_DIR",
	}
	loginPathFlag = cli.StringFlag{
		Name:   loginPath,
		Usage:  "URL to login page for this application",
		EnvVar: "LOGIN_PATH",
	}

	ldapHostFlag = cli.StringFlag{
		Name:   ldapHost,
		Usage:  "ldap hostname",
		EnvVar: "LDAP_HOST",
		Value:  "localhost",
	}
	ldapPortFlag = cli.IntFlag{
		Name:   ldapPort,
		Usage:  "ldap port",
		EnvVar: "LDAP_PORT",
		Value:  389,
	}
	ldapTLSFlag = cli.BoolFlag{
		Name:   ldapTLS,
		Usage:  "whether to use TLS with ldap",
		EnvVar: "LDAP_TLS",
	}
	ldapAdminUserFlag = cli.StringFlag{
		Name:   ldapAdminUser,
		Usage:  "ldap admin username",
		EnvVar: "LDAP_ADMIN_USER",
	}
	ldapAdminPassFlag = cli.StringFlag{
		Name:   ldapAdminPass,
		Usage:  "ldap admin password",
		EnvVar: "LDAP_ADMIN_PASS",
	}
	ldapBaseDNFlag = cli.StringFlag{
		Name:   ldapBaseDN,
		Usage:  "base DN for LDAP searches",
		EnvVar: "LDAP_BASE_DN",
	}
)
