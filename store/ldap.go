package store // import "breve.us/authsvc/store"

import (
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	ldap "gopkg.in/ldap.v2"
)

// Errors
var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNotSupported   = errors.New("not supported")
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

// Connect is a helper function for connecting to LDAP
func (c *LDAPConfig) Connect() (*ldap.Conn, error) {
	cn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.Host, c.Port))
	if err != nil {
		return nil, err
	}

	if c.UseTLS {
		if err = cn.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			cn.Close()
			return nil, err
		}
	}

	if err = cn.Bind(c.Username, c.Password); err != nil {
		cn.Close()
		return nil, err
	}
	return cn, nil
}

// NewLDAPCache returns a cache suitable for interacting with LDAP
func NewLDAPCache(config *LDAPConfig, class string, recordFn func(string, string) (interface{}, func(*ldap.Conn) error)) Cache {
	return &ldapCache{config: config, class: class, recordFn: recordFn}
}

type ldapCache struct {
	config   *LDAPConfig
	class    string
	recordFn func(string, string) (interface{}, func(*ldap.Conn) error)
}

func (c *ldapCache) Delete(key string) error                 { return ErrNotSupported }
func (c *ldapCache) Put(key string, value interface{}) error { return ErrNotSupported }
func (c *ldapCache) PutUntil(time time.Time, key string, value interface{}) error {
	return ErrNotSupported
}

func (c *ldapCache) Get(key string) (interface{}, error) {
	det, fn := c.recordFn(c.config.BaseDN, key)
	if err := c.doWithConnection(fn); err != nil {
		return nil, err
	}
	return det, nil
}

func (c *ldapCache) Keys() ([]string, error) {
	var keys []string
	if err := c.doWithConnection(func(cn *ldap.Conn) error {
		filter := fmt.Sprintf("(objectClass=%s)", c.class)
		res, err := SearchLDAP(cn, c.config.BaseDN, filter, "dn")
		if err != nil {
			return err
		}
		for _, item := range res.Entries {
			keys = append(keys, item.DN)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return keys, nil
}

// SearchLDAP wraps constructing and running a boilerplate ldap search
func SearchLDAP(cn *ldap.Conn, basedn string, filter string, attributes ...string) (*ldap.SearchResult, error) {
	r := ldap.NewSearchRequest(
		basedn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter, attributes, nil)
	return cn.Search(r)
}

func (c *ldapCache) doWithConnection(fn func(cn *ldap.Conn) error) error {
	cn, err := c.config.Connect()
	if err != nil {
		return err
	}
	defer cn.Close()

	return fn(cn)
}
