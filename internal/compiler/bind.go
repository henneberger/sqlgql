package compiler

import (
	"fmt"
	"regexp"
	"strings"
)

var nameRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
var dollarRE = regexp.MustCompile(`^\$(?:[A-Za-z_][A-Za-z0-9_]*)?\$`)

type Binding struct {
	Source, Path string
	OID          uint32
	Type         string
	Nullable     bool
}

// Bind only recognizes placeholders in SQL code, never quoted text or comments.
func Bind(src string) (string, []Binding, error) {
	return bind(src, func(index int, b Binding) string { return fmt.Sprintf("$%d", index) })
}

// BindSQLC uses named sqlc arguments so generated Go parameters retain their source.
func BindSQLC(src string) (string, []Binding, error) {
	return bind(src, func(index int, b Binding) string {
		fn := "sqlc.arg"
		if b.Source == "page" && (strings.HasPrefix(b.Path, "after.") || strings.HasPrefix(b.Path, "before.")) {
			fn = "sqlc.narg"
		}
		return fn + "(" + b.Source + "_" + strings.ReplaceAll(b.Path, ".", "_") + ")"
	})
}

func bind(src string, placeholder func(int, Binding) string) (string, []Binding, error) {
	var out strings.Builder
	var bindings []Binding
	seen := map[string]int{}
	for i := 0; i < len(src); {
		start := i
		switch {
		case strings.HasPrefix(src[i:], "--"):
			if n := strings.IndexByte(src[i:], '\n'); n >= 0 {
				i += n + 1
			} else {
				i = len(src)
			}
		case strings.HasPrefix(src[i:], "/*"):
			i += 2
			depth := 1
			for i < len(src) && depth > 0 {
				if strings.HasPrefix(src[i:], "/*") {
					depth++
					i += 2
				} else if strings.HasPrefix(src[i:], "*/") {
					depth--
					i += 2
				} else {
					i++
				}
			}
			if depth != 0 {
				return "", nil, fmt.Errorf("unterminated SQL comment")
			}
		case src[i] == '\'' || src[i] == '"':
			quote := src[i]
			escaped := quote == '\'' && i > 0 && (src[i-1] == 'E' || src[i-1] == 'e')
			i++
			closed := false
			for i < len(src) {
				if escaped && src[i] == '\\' {
					i += 2
					continue
				}
				if src[i] == quote {
					i++
					if i < len(src) && src[i] == quote {
						i++
						continue
					}
					closed = true
					break
				}
				i++
			}
			if !closed {
				return "", nil, fmt.Errorf("unterminated SQL quote")
			}
		case src[i] == '$':
			tag := dollarRE.FindString(src[i:])
			if tag != "" {
				n := strings.Index(src[i+len(tag):], tag)
				if n < 0 {
					return "", nil, fmt.Errorf("unterminated dollar quote")
				}
				i += len(tag)*2 + n
			} else {
				if i+1 < len(src) && src[i+1] >= '0' && src[i+1] <= '9' {
					return "", nil, fmt.Errorf("use {{param.name}} instead of positional SQL parameters")
				}
				i++
			}
		case strings.HasPrefix(src[i:], "{{"):
			n := strings.Index(src[i+2:], "}}")
			if n < 0 {
				return "", nil, fmt.Errorf("unterminated placeholder")
			}
			token := strings.TrimSpace(src[i+2 : i+2+n])
			parts := strings.Split(token, ".")
			if len(parts) < 2 || (parts[0] != "param" && parts[0] != "env" && parts[0] != "parent" && parts[0] != "page") {
				return "", nil, fmt.Errorf("invalid placeholder %q", token)
			}
			for _, p := range parts[1:] {
				if !nameRE.MatchString(p) {
					return "", nil, fmt.Errorf("invalid placeholder %q", token)
				}
			}
			if parts[0] == "param" && len(parts) != 2 {
				return "", nil, fmt.Errorf("param must be a single GraphQL argument name: %s", token)
			}
			index, ok := seen[token]
			if !ok {
				bindings = append(bindings, Binding{Source: parts[0], Path: strings.Join(parts[1:], ".")})
				index = len(bindings)
				seen[token] = index
			}
			out.WriteString(placeholder(index, bindings[index-1]))
			i += n + 4
			continue
		default:
			i++
		}
		out.WriteString(src[start:i])
	}
	return out.String(), bindings, nil
}
