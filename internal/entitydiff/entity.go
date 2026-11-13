package entitydiff

// EntityKind names the sort of declaration an Entity is.
type EntityKind string

const (
	// KindFunction is a top-level function.
	KindFunction EntityKind = "func"
	// KindMethod is a function with a receiver.
	KindMethod EntityKind = "method"
	// KindType is a named type declaration or type alias.
	KindType EntityKind = "type"
	// KindClass is a class declaration.
	KindClass EntityKind = "class"
	// KindInterface is a TypeScript interface declaration.
	KindInterface EntityKind = "interface"
	// KindEnum is a TypeScript enum declaration.
	KindEnum EntityKind = "enum"
)

// Entity is one declaration seen in a diff.
type Entity struct {
	Kind EntityKind
	// Name is the qualified name, e.g. "Read", "(*Store).Write" or "T.String".
	Name string
	// Signature is the declaration header with whitespace normalized,
	// e.g. "func Read(id string) (*CheckpointData, error)".
	Signature string
}

// entityKey identifies an entity within one file.
type entityKey struct {
	kind EntityKind
	name string
}

func (e Entity) key() entityKey {
	return entityKey{kind: e.Kind, name: e.Name}
}
