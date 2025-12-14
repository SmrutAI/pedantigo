package pedantigo

// ExtraFieldsMode controls how unknown JSON fields are handled during Unmarshal.
type ExtraFieldsMode int

const (
	// ExtraIgnore ignores unknown JSON fields (default behavior).
	ExtraIgnore ExtraFieldsMode = iota
	// ExtraForbid rejects JSON with unknown fields.
	ExtraForbid
	// ExtraAllow stores unknown fields (reserved for future use).
	ExtraAllow
)

// ValidatorOptions configures validator behavior.
type ValidatorOptions struct {
	// StrictMissingFields controls whether missing fields without defaults are errors
	// When true (default): missing fields without defaults cause validation errors
	// When false: missing fields are left as zero values (user handles with pointers)
	StrictMissingFields bool

	// ExtraFields controls how unknown JSON fields are handled during Unmarshal.
	// Default is ExtraIgnore (unknown fields are silently ignored).
	ExtraFields ExtraFieldsMode
}

// DefaultValidatorOptions returns the default validator options.
func DefaultValidatorOptions() ValidatorOptions {
	return ValidatorOptions{
		StrictMissingFields: true,
		ExtraFields:         ExtraIgnore,
	}
}
