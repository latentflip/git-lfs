package gitattr

import (
	"strings"

	"github.com/git-lfs/gitobj"
)

type Tree struct {
	Lines    []*Line
	Children map[string]*Tree
}

func New(db *gitobj.ObjectDatabase, t *gitobj.Tree) (*Tree, error) {
	children := make(map[string]*Tree)
	lines, err := linesInTree(db, t)
	if err != nil {
		return nil, err
	}

	for _, entry := range t.Entries {
		if entry.Type() != gitobj.TreeObjectType {
			continue
		}

		t, err := db.Tree(entry.Oid)
		if err != nil {
			return nil, err
		}

		at, err := New(db, t)
		if err != nil {
			return nil, err
		}

		children[entry.Name] = at
	}

	return &Tree{
		Lines:    lines,
		Children: children,
	}, nil
}

func linesInTree(db *gitobj.ObjectDatabase, t *gitobj.Tree) ([]*Line, error) {
	var at int = -1
	for i, e := range t.Entries {
		if e.Name == ".gitattributes" {
			at = i
			break
		}
	}

	if at < 0 {
		return nil, nil
	}

	blob, err := db.Blob(t.Entries[at].Oid)
	if err != nil {
		return nil, err
	}
	defer blob.Close()

	return ParseLines(blob.Contents)
}

func (t *Tree) Applied(to string) []*Attr {
	var attrs []*Attr
	for _, line := range t.Lines {
		if line.Pattern.Match(to) {
			attrs = append(attrs, line.Attrs...)
		}
	}

	splits := strings.SplitN(to, "/", 2)
	if len(splits) == 2 {
		car, cdr := splits[0], splits[1]
		if child, ok := t.Children[car]; ok {
			attrs = append(attrs, child.Applied(cdr)...)
		}
	}

	return attrs
}
