package gitattr

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/git-lfs/git-lfs/errors"
	"github.com/git-lfs/wildmatch"
)

type Line struct {
	Pattern *wildmatch.Wildmatch
	Attrs   []*Attr
}

type Attr struct {
	K           string
	V           string
	Unspecified bool
}

func ParseLines(r io.Reader) ([]*Line, error) {
	var lines []*Line

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {

		text := strings.TrimSpace(scanner.Text())
		if len(text) == 0 {
			continue
		}

		var pattern string
		var applied string

		switch text[0] {
		case '#':
			continue
		case '"':
			var err error
			last := strings.LastIndex(text, "#")
			if last < 0 {
				return nil, errors.Errorf("git/gitattr: unbalanced quote: %s", text)
			}
			pattern, err = strconv.Unquote(text[:last])
			if err != nil {
				return nil, errors.Wrapf(err, "git/gitattr")
			}
			applied = text[last:]
		default:
			splits := strings.SplitN(text, " ", 2)
			if len(splits) != 2 {
				return nil, errors.Errorf("git/gitattr: malformed line: %s", text)
			}

			pattern = splits[0]
			applied = splits[1]
		}

		var attrs []*Attr

		for _, s := range strings.Split(applied, " ") {
			var attr Attr

			if strings.HasPrefix(s, "-") {
				attr.K = strings.TrimPrefix(s, "-")
				attr.V = "false"
			} else if strings.HasPrefix(s, "!") {
				attr.K = strings.TrimPrefix(s, "!")
				attr.Unspecified = true
			} else {
				splits := strings.SplitN(s, "=", 2)
				if len(splits) != 2 {
					return nil, errors.Errorf("git/gitattr: malformed attribute: %s", s)
				}
				attr.K = splits[0]
				attr.V = splits[1]
			}

			attrs = append(attrs, &attr)
		}

		lines = append(lines, &Line{
			Pattern: wildmatch.NewWildmatch(
				pattern,
				wildmatch.SystemCase,
			),
			Attrs: attrs,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
