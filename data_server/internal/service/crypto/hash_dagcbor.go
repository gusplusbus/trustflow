package crypto

import (
	"crypto/sha256"
	"time"

	cbor "github.com/fxamacker/cbor/v2"
)

var enc cbor.EncMode

func init() {
	// Use broadly supported options so older fxamacker/cbor versions compile.
	// Canonical map key sorting + RFC3339 time gives us deterministic encoding.
	opts := cbor.EncOptions{
		Sort: cbor.SortCanonical,   // widely available (fallback if SortCoreDeterministic isn't present)
		Time: cbor.TimeRFC3339,
		// If your version supports these and you want them, you can re-add later:
		// IndefLength:  cbor.IndefLengthForbidden,
		// ShortestFloat: cbor.ShortestFloat16,
		// NaNConvert:   cbor.NaNConvert7e00,
		// CTAP2Canonical: true, // remove for compatibility
	}
	var err error
	enc, err = opts.EncMode()
	if err != nil {
		panic(err)
	}
}

// CanonItem is the deterministic shape we hash.
type CanonItem struct {
	Provider        string         `cbor:"provider"`
	ProviderEventID string         `cbor:"provider_event_id"`
	IssueNodeID     string         `cbor:"issue_node_id,omitempty"`
	Type            string         `cbor:"type"`
	Actor           *string        `cbor:"actor,omitempty"`
	CreatedAt       time.Time      `cbor:"created_at"` // RFC3339 via EncOptions
	Payload         map[string]any `cbor:"payload"`
}

// HashDAGCBOR encodes v using canonical CBOR and returns (encoded, sha256).
func HashDAGCBOR(v any) ([]byte, []byte, error) {
	b, err := enc.Marshal(v)
	if err != nil {
		return nil, nil, err
	}
	sum := sha256.Sum256(b)
	return b, sum[:], nil
}

