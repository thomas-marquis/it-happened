package eventest

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMarbleSyntax = errors.New("invalid marble syntax")
	ErrEmptyMarble  = errors.Join(ErrMarbleSyntax, errors.New("empty marble"))
)

func parseMarble(marble string, defaultIckDuration time.Duration) ([]op, error) {
	var parsed []op
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
			parsed = append(parsed, startEventOp{})
			i++
		case c == '-':
			parsed = append(parsed, waitOp{Duration: defaultIckDuration})
			i++
		case c == '_':
			for i < len(marble) && marble[i] == '_' {
				i++
			}
			parsed = append(parsed, waitOp{Duration: defaultIckDuration})
		case (c == '/') || ((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')):
			label := parseLabel(marble, &i)
			if i < len(marble)-2 && marble[i:i+2] == "<-" {
				i += 2
				fLabel := parseLabel(marble, &i)
				parsed = append(parsed, eventWithFollowupOp{EventName: label, From: fLabel})
			} else {
				parsed = append(parsed, eventOp{Name: label})
			}
		case c == '(':
			grp, err := parseGroup(marble, &i, '(', ')', "unbalanced parenthesis", defaultIckDuration)
			if err != nil {
				return nil, err
			}

			parsed = append(parsed, unorderedGroupOp{Ops: grp})
		case c == '[':
			grp, err := parseGroup(marble, &i, '[', ']', "squared brackets", defaultIckDuration)
			if err != nil {
				return nil, err
			}

			parsed = append(parsed, orderedGroupOp{Ops: grp})
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

func parseGroup(marble string, i *int, open, close rune, errMsg string, defaultIckDuration time.Duration) ([]op, error) {
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
	grpOps, err := parseMarble(grpMarble, defaultIckDuration)
	if err != nil {
		return nil, err
	}
	return grpOps, nil
}
