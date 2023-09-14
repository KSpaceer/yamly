package yayamls

// Marshaler interface can be implemented to customize type's behaviour when being
// marshaled into a YAML document returning raw representation
type Marshaler interface {
	MarshalYAML() ([]byte, error)
}
