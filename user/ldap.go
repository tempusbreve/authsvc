package user // import "breve.us/authsvc/user"

import (
	"errors"
	"fmt"
	"hash/fnv"

	ldap "gopkg.in/ldap.v2"

	"breve.us/authsvc/common"
	"breve.us/authsvc/store"
)

// Errors
var (
	ErrNotFound = errors.New("not found")
)

// NewLDAPChecker returns a password checker using LDAP
func NewLDAPChecker(config *store.LDAPConfig) common.PasswordChecker {
	return &checker{cfg: config}
}

type checker struct {
	cfg *store.LDAPConfig
}

func (c *checker) IsAuthenticated(username string, password string) bool {
	cn, err := c.cfg.Connect()
	if err != nil {
		return false
	}
	defer cn.Close()

	var res *ldap.SearchResult
	for _, attr := range usernameAttributes {
		if res, err = store.SearchLDAP(cn, c.cfg.BaseDN, fmt.Sprintf("(%s=%s)", attr, username), "cn"); err != nil {
			// TODO: ?
			continue
		}
		switch len(res.Entries) {
		case 1:
			e := res.Entries[0]
			if err = cn.Bind(e.DN, password); err == nil {
				// TODO: ?
				return true
			}
		case 0:
			continue
		default:
			// TODO: ?
			continue
		}
	}
	return false
}

// NewLDAPCache returns a cache suitable for interacting with LDAP
func NewLDAPCache(config *store.LDAPConfig) store.Cache {
	return store.NewLDAPCache(config, "inetOrgPerson", recordFn)
}

var usernameAttributes = []string{"uid", "mail"}

func recordFn(basedn string, key string) (interface{}, func(*ldap.Conn) error) {
	det := &Details{}
	fn := func(cn *ldap.Conn) error {
		var (
			err error
			res *ldap.SearchResult
		)
		for _, attr := range usernameAttributes {
			if res, err = store.SearchLDAP(cn, basedn, fmt.Sprintf("(%s=%s)", attr, key), "*"); err != nil {
				// TODO: ?
				continue
			}
			switch len(res.Entries) {
			case 0:
				continue
			case 1:
				return populateDetails(attr, det, res.Entries[0])
			default:
				// TODO: ?
				continue
			}
		}
		return ErrNotFound
	}
	return det, fn
}

func populateDetails(key string, d *Details, m *ldap.Entry) error {
	// Generate hash as a probably-unique ID placeholder
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
