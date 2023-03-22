package rdns

import (
	"testing"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
)

func TestDomainDB(t *testing.T) {
	loader := NewStaticLoader([]string{
		"domain1.com.",    // exact match
		".domain2.com.",   // exact match and subdomains
		"x.domain2.com",   // above rule should take precendence
		"*.domain3.com",   // subdomains only
		"x.x.domain3.com", // more general wildcard above should take precedence
		"domain4.com",     // the more general rule below wins
		".domain4.com",
	})

	db := NewDomainDB("testlist", loader)
	m, err := db.Reload()
	require.NoError(t, err)

	tests := []struct {
		q     string
		match bool
	}{
		// exact
		{"domain1.com.", true},
		{"x.domain1.com.", false},

		// exact and subdomains
		{"domain2.com.", true},
		{"sub.domain2.com.", true},

		// wildcard (match only on subdomains)
		{"domain3.com.", false},
		{"sub.domain3.com.", true},

		// two rules for this, the generic one wins
		{"domain4.com.", true},
		{"sub.domain4.com.", true},

		// not matching
		{"unblocked.test.", false},
		{"com.", false},
	}
	for _, test := range tests {
		q := dns.Question{Name: test.q, Qtype: dns.TypeA, Qclass: dns.ClassINET}
		_, _, _, ok := m.Match(q)
		require.Equal(t, test.match, ok, "query: %s", test.q)
	}
}

func TestDomainDBError(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"sub.*.com"},
		{"*domain.com"},
	}
	for _, test := range tests {
		loader := NewStaticLoader([]string{test.name})
		db := NewDomainDB("testlist", loader)
		_, err := db.Reload()
		require.Error(t, err)
	}
}
