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
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			parsed = append(parsed, eventOp{Name: string(c)})
			i++
		case c == '/':
			lb := strings.Builder{}
			i++
			for i < len(marble) {
				c = marble[i]
				if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') {
					lb.WriteByte(marble[i])
					i++
				} else {
					break
				}
			}
			parsed = append(parsed, eventOp{Name: lb.String()})
		case c == '(':
			var end int
			cnt := 1
			i++
			start := i
			for _, x := range marble[i:] {
				if x == ')' {
					cnt--
				} else if x == '(' {
					cnt++
				}
				if cnt == 0 {
					end = i
					i++
					break
				}
				i++
			}

			if end == 0 {
				return nil, errors.Join(
					ErrMarbleSyntax,
					fmt.Errorf("unbalanced parenthesis at %d", start-1),
				)
			}

			grpMarble := marble[start:end]
			grpOps, err := parseMarble(grpMarble, defaultIckDuration)
			if err != nil {
				return nil, err
			}
			parsed = append(parsed, unorderedGroupOp{Ops: grpOps})
		case c == '[':
			var end int
			cnt := 1
			i++
			start := i
			for _, x := range marble[i:] {
				if x == ']' {
					cnt--
				} else if x == '[' {
					cnt++
				}
				if cnt == 0 {
					end = i
					i++
					break
				}
				i++
			}

			if end == 0 {
				return nil, errors.Join(
					ErrMarbleSyntax,
					fmt.Errorf("unbalanced parenthesis at %d", start-1),
				)
			}

			grpMarble := marble[start:end]
			grpOps, err := parseMarble(grpMarble, defaultIckDuration)
			if err != nil {
				return nil, err
			}
			parsed = append(parsed, orderedGroupOp{Ops: grpOps})
		default:
			return nil, errors.Join(
				ErrMarbleSyntax,
				fmt.Errorf("unexpected character %q at %d", c, i),
			)
		}
	}

	return parsed, nil
}
