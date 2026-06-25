package signature

import (
	"hash/fnv"
	"math"
	"sort"
	"strings"
)

// FeatureVector represents a normalized error signature for matching.
type FeatureVector struct {
	ErrorType        string
	ErrorMessageHash uint64
	LocationPattern  string
	StackHashes      []uint64
	ChangeType       string
	DistanceToError  int // 0=same file, 1=same module, 2=different module
	IsNewInDiff      bool
}

// Signature is a known failure pattern with diagnosis and resolution.
type Signature struct {
	ID          string       `json:"signature_id"`
	CreatedAt   string       `json:"created_at"`
	MatchCount  int          `json:"match_count"`
	Features    FeatureVector `json:"feature_vector"`
	Diagnosis   Diagnosis    `json:"diagnosis"`
	Resolution  Resolution   `json:"resolution"`
}

// Diagnosis is the root cause analysis.
type Diagnosis struct {
	RootCause         string  `json:"root_cause"`
	RootCauseCategory string  `json:"root_cause_category"`
	Reliability       float64 `json:"reliability"`
}

// Resolution is the fix strategy.
type Resolution struct {
	Strategy             string   `json:"strategy"`
	FixTemplate          string   `json:"fix_template"`
	VerificationRequired []string `json:"verification_required"`
}

// MatchResult is a matched signature with similarity score.
type MatchResult struct {
	SignatureID       string  `json:"signature_id"`
	Similarity        float64 `json:"similarity"`
	Reliability       float64 `json:"reliability"`
	MatchCount        int     `json:"match_count"`
	ResolutionStrategy string  `json:"resolution_strategy"`
	FixTemplate       string  `json:"fix_template"`
}

// MatcherConfig holds weights for the similarity calculation.
type MatcherConfig struct {
	WErrorType     float64
	WErrorMessage  float64
	WLocation      float64
	WStackTrace    float64
	WChangeContext float64
	WTemporal      float64
	MinConfidence  float64
	TopK           int
}

// DefaultMatcherConfig returns sensible defaults.
func DefaultMatcherConfig() MatcherConfig {
	return MatcherConfig{
		WErrorType:     0.30,
		WErrorMessage:  0.15,
		WLocation:      0.25,
		WStackTrace:    0.15,
		WChangeContext: 0.10,
		WTemporal:      0.05,
		MinConfidence:  0.65,
		TopK:           3,
	}
}

// Matcher performs weighted similarity matching against a signature database.
type Matcher struct {
	config     MatcherConfig
	signatures []*Signature
}

// NewMatcher creates a new matcher with the given config.
func NewMatcher(config MatcherConfig) *Matcher {
	return &Matcher{
		config:     config,
		signatures: make([]*Signature, 0),
	}
}

// Load loads signatures into the matcher.
func (m *Matcher) Load(signatures []*Signature) {
	m.signatures = signatures
}

// Match finds the top-K matching signatures for a feature vector.
func (m *Matcher) Match(query *FeatureVector) []MatchResult {
	var results []MatchResult

	for _, sig := range m.signatures {
		score := m.similarity(query, &sig.Features)

		if score >= m.config.MinConfidence {
			results = append(results, MatchResult{
				SignatureID:        sig.ID,
				Similarity:         score,
				Reliability:        sig.Diagnosis.Reliability,
				MatchCount:         sig.MatchCount,
				ResolutionStrategy: sig.Resolution.Strategy,
				FixTemplate:        sig.Resolution.FixTemplate,
			})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Similarity > results[j].Similarity
	})

	if len(results) > m.config.TopK {
		results = results[:m.config.TopK]
	}

	return results
}

// similarity computes weighted similarity between a query and a signature.
func (m *Matcher) similarity(query, sig *FeatureVector) float64 {
	score := 0.0

	// Error type match (exact)
	if strings.EqualFold(query.ErrorType, sig.ErrorType) && query.ErrorType != "" {
		score += m.config.WErrorType
	}

	// Error message hash similarity
	if query.ErrorMessageHash != 0 && sig.ErrorMessageHash != 0 {
		if query.ErrorMessageHash == sig.ErrorMessageHash {
			score += m.config.WErrorMessage
		}
	}

	// Location similarity
	score += m.config.WLocation * locationSimilarity(query.LocationPattern, sig.LocationPattern)

	// Stack trace similarity (Jaccard)
	score += m.config.WStackTrace * jaccardSimilarity(query.StackHashes, sig.StackHashes)

	// Change context similarity
	if query.ChangeType == sig.ChangeType && query.ChangeType != "" {
		score += m.config.WChangeContext * 0.5
	}
	if query.DistanceToError == sig.DistanceToError {
		score += m.config.WChangeContext * 0.5
	}

	// Temporal bonus (new errors matching recent patterns get a boost)
	if query.IsNewInDiff {
		score += m.config.WTemporal * 0.5
	}

	return math.Min(score, 1.0)
}

// locationSimilarity computes how similar two file paths are.
func locationSimilarity(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1.0
	}
	// Check common prefix
	aLower := strings.ToLower(a)
	bLower := strings.ToLower(b)
	if aLower == bLower {
		return 1.0
	}
	// Check if same directory
	aParts := strings.Split(aLower, "/")
	bParts := strings.Split(bLower, "/")
	if len(aParts) > 1 && len(bParts) > 1 {
		aDir := strings.Join(aParts[:len(aParts)-1], "/")
		bDir := strings.Join(bParts[:len(bParts)-1], "/")
		if aDir == bDir {
			return 0.8
		}
	}
	// Check file name similarity
	if len(aParts) > 0 && len(bParts) > 0 {
		aFile := aParts[len(aParts)-1]
		bFile := bParts[len(bParts)-1]
		if aFile == bFile {
			return 0.6
		}
	}
	return 0
}

// jaccardSimilarity computes Jaccard similarity between two sets of hashes.
func jaccardSimilarity(a, b []uint64) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 0
	}
	setA := make(map[uint64]bool)
	setB := make(map[uint64]bool)
	for _, h := range a {
		setA[h] = true
	}
	for _, h := range b {
		setB[h] = true
	}
	intersection := 0
	for h := range setA {
		if setB[h] {
			intersection++
		}
	}
	union := len(setA) + len(setB) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// HashString returns a FNV-1a 64-bit hash of a string.
func HashString(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}
