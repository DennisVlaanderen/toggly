package store

// Flag is a single feature flag record replicated across the cluster.
type Flag struct {
	Key     string `json:"key"`
	Enabled bool   `json:"enabled"`
	Value   string `json:"value,omitempty"`
	Version uint64 `json:"version"`
}
