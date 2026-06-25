package signature

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// DB manages the signature database.
type DB struct {
	path       string
	matcher    *Matcher
	signatures []*Signature
	autoLearn  bool
}

// NewDB creates a new signature database.
func NewDB(path string, seedFile string, autoLearn bool) (*DB, error) {
	db := &DB{
		path:      path,
		matcher:   NewMatcher(DefaultMatcherConfig()),
		autoLearn: autoLearn,
	}

	// Load seed signatures
	if seedFile != "" {
		if data, err := os.ReadFile(seedFile); err == nil {
			var seeds map[string][]*Signature
			if err := json.Unmarshal(data, &seeds); err == nil {
				db.signatures = seeds["signatures"]
				db.matcher.Load(db.signatures)
			}
		}
	}

	// Load existing signatures
	if path != "" {
		if data, err := os.ReadFile(path); err == nil {
			var sigs []*Signature
			if err := json.Unmarshal(data, &sigs); err == nil {
				db.signatures = append(db.signatures, sigs...)
				db.matcher.Load(db.signatures)
			}
		}
	}

	return db, nil
}

// Match finds matching signatures for a query.
func (db *DB) Match(query *FeatureVector) []MatchResult {
	return db.matcher.Match(query)
}

// Learn creates a new signature from a resolved escalation.
func (db *DB) Learn(features *FeatureVector, diagnosis Diagnosis, resolution Resolution) *Signature {
	sig := &Signature{
		ID:        fmt.Sprintf("sig-%s-%06d", time.Now().UTC().Format("20060102"), len(db.signatures)+1),
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Features:  *features,
		Diagnosis: diagnosis,
		Resolution: resolution,
	}
	db.signatures = append(db.signatures, sig)
	db.matcher.Load(db.signatures)

	// Persist
	if db.path != "" {
		data, _ := json.MarshalIndent(db.signatures, "", "  ")
		os.WriteFile(db.path, data, 0644)
	}

	return sig
}

// Save persists the signature database.
func (db *DB) Save() error {
	if db.path == "" {
		return nil
	}
	data, err := json.MarshalIndent(db.signatures, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(db.path, data, 0644)
}

// Count returns the number of signatures.
func (db *DB) Count() int {
	return len(db.signatures)
}

// AllSignatures returns all signatures.
func (db *DB) AllSignatures() []*Signature {
	return db.signatures
}
