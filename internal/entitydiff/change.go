package entitydiff

// ChangeKind is the kind of structural change made to an entity.
type ChangeKind string

const (
	// Added means the entity exists only after the change.
	Added ChangeKind = "added"
	// Removed means the entity exists only before the change.
	Removed ChangeKind = "removed"
	// SignatureChanged means the entity exists on both sides with a
	// different declaration header.
	SignatureChanged ChangeKind = "signature changed"
)

// Change is one entity-level change. Old is the zero Entity for Added and
// New is the zero Entity for Removed.
type Change struct {
	Kind ChangeKind
	Old  Entity
	New  Entity
}

// FileChanges holds the entity changes of one file, in source order.
// OldPath is empty for a new file and NewPath is empty for a deleted file.
type FileChanges struct {
	OldPath string
	NewPath string
	Changes []Change
}

// DisplayPath is the file's path for display: "old → new" for a renamed
// file, otherwise whichever path exists.
func (f FileChanges) DisplayPath() string {
	switch {
	case f.OldPath == "":
		return f.NewPath
	case f.NewPath == "", f.OldPath == f.NewPath:
		return f.OldPath
	default:
		return f.OldPath + " → " + f.NewPath
	}
}
