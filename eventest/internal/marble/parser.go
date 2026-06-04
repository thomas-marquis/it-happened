package marble

import (
	"errors"
	"fmt"
	"strings"
)

var (
	ErrMarbleSyntax = errors.New("invalid marble syntax")
	ErrEmptyMarble  = errors.Join(ErrMarbleSyntax, errors.New("empty marble"))
)

func Parse(marble string) ([]Op, error) {
	var pos int
	return parse(marble, &pos)
}

func parse(marble string, pos *int) ([]Op, error) {
	var (
		parsed []Op
	)
	if marble == "" {
		return nil, ErrEmptyMarble
	}

	i := 0
	for i < len(marble) {
		c := marble[i]
		switch {
		case c == ' ', c == '\t', c == '\n', c == '\r':
			i++
		case c == '^':
			if i != 0 {
				return nil, errors.Join(
					ErrMarbleSyntax,
					fmt.Errorf("unexpected ^ at %d", i),
				)
			}
			parsed = append(parsed, StartEventOp{})
			*pos++
			i++
		case c == '-':
			parsed = append(parsed, WaitOp{})
			*pos++
			i++
		case c == '_':
			for i < len(marble) && marble[i] == '_' {
				i++
			}
			parsed = append(parsed, WaitOp{})
			*pos++
		case (c == '/') || ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')):
			label := parseLabel(marble, &i)
			if i < len(marble)-2 && marble[i:i+2] == "<-" {
				i += 2
				fLabel := parseLabel(marble, &i)
				parsed = append(parsed, EventWithFollowupOp{EventName: label, From: fLabel})
			} else {
				parsed = append(parsed, EventOp{Name: label})
			}
			*pos++
		case c == '(':
			startPos := *pos
			*pos++
			grp, err := parseGroup(marble, &i, '(', ')', "unbalanced parenthesis", pos)
			if err != nil {
				return nil, err
			}

			parsed = append(parsed, UnorderedGroupStartOp{EndPos: *pos})
			parsed = append(parsed, grp...)
			parsed = append(parsed, UnorderedGroupEndOp{StartPos: startPos})
			*pos++
		case c == '[':
			startPos := *pos
			*pos++
			grp, err := parseGroup(marble, &i, '[', ']', "squared brackets", pos)
			if err != nil {
				return nil, err
			}

			parsed = append(parsed, OrderedGroupStartOp{EndPos: *pos})
			parsed = append(parsed, grp...)
			parsed = append(parsed, OrderedGroupEndOp{StartPos: startPos})
			*pos++
		default:
			return nil, errors.Join(
				ErrMarbleSyntax,
				fmt.Errorf("unexpected character %q at %d", c, i),
			)
		}
	}

	return parsed, nil
}

func parseLabel(marble string, i *int) string {
	c := marble[*i]
	var label string
	if c == '/' {
		lb := strings.Builder{}
		*i++
		for *i < len(marble) {
			c = marble[*i]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				lb.WriteByte(marble[*i])
				*i++
			} else {
				break
			}
		}
		label = lb.String()
	} else {
		label = string(c)
		*i++
	}
	return label
}

func parseGroup(marble string, i *int, open, close rune, errMsg string, pos *int) ([]Op, error) {
	var end int
	cnt := 1
	*i++
	start := *i
	for _, x := range marble[*i:] {
		if x == close {
			cnt--
		} else if x == open {
			cnt++
		}
		if cnt == 0 {
			end = *i
			*i++
			break
		}
		*i++
	}

	if end == 0 {
		return nil, errors.Join(
			ErrMarbleSyntax,
			fmt.Errorf("%s at %d", errMsg, start-1),
		)
	}

	grpMarble := marble[start:end]
	grpOps, err := parse(grpMarble, pos)
	if err != nil {
		return nil, err
	}
	return grpOps, nil
}
