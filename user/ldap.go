package user // import "breve.us/authsvc/user"

import (
	"crypto/tls"
	"fmt"
	"log"
	"time"

	ldap "gopkg.in/ldap.v2"

	"breve.us/authsvc/store"
)

// LDAPConfig describes connection details to an LDAP server
type LDAPConfig struct {
	Host     string
	Port     int
	UseTLS   bool
	Username string
	Password string
	BaseDN   string
}

// NewLDAPCache returns a cache suitable for interacting with LDAP
func NewLDAPCache(config *LDAPConfig) store.Cache {
	return &ldapCache{config: config, class: "inetOrgPerson"}
}

type ldapCache struct {
	config *LDAPConfig
	class  string
}

func (c *ldapCache) Put(key string, value interface{}) error                      { return nil }
func (c *ldapCache) PutUntil(time time.Time, key string, value interface{}) error { return nil }
func (c *ldapCache) Get(key string) (interface{}, error)                          { return nil, nil }
func (c *ldapCache) Delete(key string) error                                      { return nil }

func (c *ldapCache) Keys() ([]string, error) {
	cn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port))
	if err != nil {
		log.Printf("ERROR: dialing LDAP: %v", err)
		return nil, err
	}
	defer cn.Close()

	if c.config.UseTLS {
		if err = cn.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			log.Printf("ERROR: StartTLS: %v", err)
			return nil, err
		}
	}

	if err = cn.Bind(c.config.Username, c.config.Password); err != nil {
		log.Printf("ERROR: LDAP Bind %q/%q: %v", c.config.Username, c.config.Password, err)
		return nil, err
	}

	r := ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(objectClass=%s)", c.class),
		[]string{"dn", "cn", "mail"},
		nil,
	)

	res, err := cn.Search(r)
	if err != nil {
		log.Printf("ERROR: listing keys: %v", err)
		return nil, err
	}

	var keys []string
	for _, item := range res.Entries {
		keys = append(keys, item.DN)
	}

	return keys, nil
}
