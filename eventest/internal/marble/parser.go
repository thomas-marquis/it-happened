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
	node, err := ParseAsNode(marble)
	if err != nil {
		return nil, err
	}
	return ToOpList(node), nil
}

func ParseAsNode(marble string) (Node, error) {
	if marble == "" {
		return nil, ErrEmptyMarble
	}
	var pos int
	return parse(marble, &pos)
}

func parse(marble string, pos *int) (Node, error) {
	var children []Node

	for *pos < len(marble) {
		c := marble[*pos]
		switch {
		case c == ' ', c == '\t', c == '\n', c == '\r':
			*pos++
		case c == '^':
			children = append(children, &PlaceholderNode{pos: Position{Offset: *pos}})
			*pos++
		case c == '-':
			children = append(children, &WaitNode{pos: Position{Offset: *pos}})
			*pos++
		case c == '_':
			start := *pos
			for *pos < len(marble) && marble[*pos] == '_' {
				*pos++
			}
			children = append(children, &WaitNode{pos: Position{Offset: start}})
		case (c == '/') || ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')):
			start := *pos
			label := parseLabel(marble, pos)
			if *pos < len(marble)-1 && marble[*pos:*pos+2] == "<-" {
				*pos += 2
				fLabel := parseLabel(marble, pos)
				children = append(children, &FollowupNode{
					NewEvent: label,
					OfEvent:  fLabel,
					pos:      Position{Offset: start},
				})
			} else {
				children = append(children, &EventNode{
					Name: label,
					pos:  Position{Offset: start},
				})
			}
		case c == '(':
			group, err := parseGroupNode(marble, pos, '(', ')', "unbalanced parenthesis", false)
			if err != nil {
				return nil, err
			}
			children = append(children, group)
		case c == '[':
			group, err := parseGroupNode(marble, pos, '[', ']', "unbalanced squared brackets", true)
			if err != nil {
				return nil, err
			}
			children = append(children, group)
		default:
			return nil, errors.Join(
				ErrMarbleSyntax,
				fmt.Errorf("unexpected character %q at %d", c, *pos),
			)
		}
	}

	return &SequenceNode{Children: children}, nil
}

func parseLabel(marble string, pos *int) string {
	c := marble[*pos]
	var label string
	if c == '/' {
		lb := strings.Builder{}
		*pos++
		for *pos < len(marble) {
			c = marble[*pos]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
				lb.WriteByte(marble[*pos])
				*pos++
			} else {
				break
			}
		}
		label = lb.String()
	} else {
		label = string(c)
		*pos++
	}
	return label
}

func parseGroupNode(marble string, pos *int, open, close rune, errMsg string, ordered bool) (Node, error) {
	startPos := *pos
	*pos++ // skip open char
	start := *pos

	cnt := 1
	var end int
	for *pos < len(marble) {
		x := rune(marble[*pos])
		if x == open {
			cnt++
		} else if x == close {
			cnt--
		}

		if cnt == 0 {
			end = *pos
			*pos++ // skip close char
			break
		}
		*pos++
	}

	if cnt != 0 {
		return nil, errors.Join(
			ErrMarbleSyntax,
			fmt.Errorf("%s at %d", errMsg, start-1),
		)
	}

	grpMarble := marble[start:end]
	var grpPos int
	grpNode, err := parse(grpMarble, &grpPos)
	if err != nil {
		return nil, err
	}

	seq := grpNode.(*SequenceNode)
	return &GroupNode{
		Children: seq.Children,
		Ordered:  ordered,
		pos:      Position{Offset: startPos},
	}, nil
}
