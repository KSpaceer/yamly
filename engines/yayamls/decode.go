package yayamls

// Unmarshaler interface can be implemented to customize type's behaviour when being
// unmarshaled from YAML document using raw YAML.
type Unmarshaler interface {
	UnmarshalYAML([]byte) error
}
