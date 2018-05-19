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
func NewLDAPCache(config *LDAPConfig, class string) store.Cache {
	return &ldapCache{config: config, class: class}
}

type ldapCache struct {
	config *LDAPConfig
	class  string
}

func (c *ldapCache) Put(key string, value interface{}) error                      { return nil }
func (c *ldapCache) PutUntil(time time.Time, key string, value interface{}) error { return nil }
func (c *ldapCache) Get(key string) (interface{}, error)                          { return nil, nil }
func (c *ldapCache) Delete(key string) error                                      { return nil }

func (c *ldapCache) Keys() []string {
	cn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port))
	if err != nil {
		return nil
	}
	defer cn.Close()

	if c.config.UseTLS {
		if err = cn.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			return nil
		}
	}

	r := ldap.NewSearchRequest(
		c.config.BaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		fmt.Sprintf("(&(objectClass=%s))", c.class),
		[]string{"dn"},
		nil,
	)

	res, err := cn.Search(r)
	if err != nil {
		log.Printf("ERROR: listing keys: %v", err)
		return nil
	}

	var keys []string
	for _, item := range res.Entries {
		keys = append(keys, item.DN)
	}

	return keys
}
