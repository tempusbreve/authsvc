package user // import "breve.us/authsvc/user"

import (
	"crypto/tls"
	"errors"
	"fmt"
	"hash/fnv"
	"time"

	ldap "gopkg.in/ldap.v2"

	"breve.us/authsvc/store"
)

// Errors
var (
	ErrNotImplemented = errors.New("not implemented")
	ErrNotSupported   = errors.New("not supported")
	ErrNotFound       = errors.New("not found")
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

func (c *ldapCache) Delete(key string) error                 { return ErrNotSupported }
func (c *ldapCache) Put(key string, value interface{}) error { return ErrNotSupported }
func (c *ldapCache) PutUntil(time time.Time, key string, value interface{}) error {
	return ErrNotSupported
}

var usernameAttributes = []string{"uid", "mail"}

func (c *ldapCache) Get(key string) (interface{}, error) {
	det := &Details{}
	if err := c.doWithConnection(func(cn *ldap.Conn) error {
		var (
			err error
			res *ldap.SearchResult
		)
		for _, attr := range usernameAttributes {
			if res, err = search(cn, c.config.BaseDN, fmt.Sprintf("(%s=%s)", attr, key), "*"); err != nil {
				continue
			}
			if len(res.Entries) != 1 {
				// TODO: debug logging?
				continue
			}
			return populateDetails(attr, det, res.Entries[0])
		}
		return ErrNotFound
	}); err != nil {
		return nil, err
	}
	return det, nil
}

func (c *ldapCache) Keys() ([]string, error) {
	var keys []string
	if err := c.doWithConnection(func(cn *ldap.Conn) error {
		filter := fmt.Sprintf("(objectClass=%s)", c.class)
		res, err := search(cn, c.config.BaseDN, filter, "dn")
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

func (c *ldapCache) doWithConnection(fn func(cn *ldap.Conn) error) error {
	cn, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", c.config.Host, c.config.Port))
	if err != nil {
		return err
	}
	defer cn.Close()

	if c.config.UseTLS {
		if err = cn.StartTLS(&tls.Config{InsecureSkipVerify: false}); err != nil {
			return err
		}
	}

	if err = cn.Bind(c.config.Username, c.config.Password); err != nil {
		return err
	}

	return fn(cn)
}

func search(cn *ldap.Conn, basedn string, filter string, attributes ...string) (*ldap.SearchResult, error) {
	r := ldap.NewSearchRequest(
		basedn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0, 0, false,
		filter, attributes, nil)
	return cn.Search(r)
}

func populateDetails(key string, d *Details, m *ldap.Entry) error {
	// Generate hash as a semi-unique ID placeholder
	h := fnv.New64a()
	fmt.Fprintf(h, m.DN)
	d.ID = h.Sum64()

	d.Username = m.GetAttributeValue(key)
	d.Password = m.GetAttributeValue("userPassword")
	d.Email = m.GetAttributeValue("mail")
	d.Name = m.GetAttributeValue("cn")
	d.State = Active
	return nil
}
