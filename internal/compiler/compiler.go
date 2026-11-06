package compiler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
)

type Column struct{ Name, Type string }
type Query struct {
	Path, Field, Type, Func, SQL, SQLC, Source string
	One                                        bool
	Batch                                      map[string]string
	Connection                                 []string
	Bindings                                   []Binding
	Columns                                    []Column
	Children                                   []*Query `json:"-"`
}

func upper(s string) string { return strings.ToUpper(s[:1]) + s[1:] }
func scalar(oid uint32) (string, error) {
	if elem, ok := arrayElement(oid); ok {
		typ, err := scalar(elem)
		if err != nil {
			return "", err
		}
		if typ == "JSON" {
			return "[JSON]", nil
		}
		return "[" + typ + "]", nil
	}
	switch oid {
	case 16:
		return "Boolean", nil
	case 20:
		return "BigInt", nil
	case 21, 23:
		return "Int", nil
	case 700, 701:
		return "Float", nil
	case 2950:
		return "UUID", nil
	case 114, 3802:
		return "JSON", nil
	case 25, 1042, 1043:
		return "String", nil
	}
	return "", fmt.Errorf("unsupported PostgreSQL type OID %d: cast the SQL expression (e.g. numeric/date to text) to choose its API representation", oid)
}

func Inspect(ctx context.Context, conn *pgx.Conn, dir string) ([]*Query, error) {
	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no .sql files in %s", dir)
	}
	var queries []*Query
	paths := map[string]*Query{}
	types := map[string]string{"Query": "reserved", "UUID": "reserved scalar", "JSON": "reserved scalar", "BigInt": "reserved scalar"}
	funcs := map[string]bool{}
	for _, file := range files {
		path := strings.TrimSuffix(filepath.Base(file), ".sql")
		parts := strings.Split(path, ".")
		if len(parts) < 2 || parts[0] != "Query" {
			return nil, fmt.Errorf("%s: expected Query.field[.nestedField].sql", file)
		}
		for _, p := range parts {
			if !nameRE.MatchString(p) || strings.HasPrefix(p, "__") {
				return nil, fmt.Errorf("invalid GraphQL name %q", p)
			}
		}
		q := &Query{Path: path, Field: parts[len(parts)-1]}
		q.Type = upper(q.Field)
		for _, p := range parts[1:] {
			q.Func += upper(p)
		}
		if prev, ok := types[q.Type]; ok {
			return nil, fmt.Errorf("type %s collides between %s and %s; use distinct field names", q.Type, prev, path)
		}
		types[q.Type] = path
		if funcs[q.Func] {
			return nil, fmt.Errorf("retriever name collision %s", q.Func)
		}
		funcs[q.Func] = true
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		for _, line := range strings.Split(string(data), "\n") {
			line = strings.TrimSpace(line)
			if line == "-- @one" {
				q.One = true
			}
			if strings.HasPrefix(line, "-- @connection ") {
				if q.Connection != nil {
					return nil, fmt.Errorf("%s: duplicate connection declaration", file)
				}
				for _, key := range strings.Split(strings.TrimPrefix(line, "-- @connection "), ",") {
					key = strings.TrimSpace(key)
					if !nameRE.MatchString(key) {
						return nil, fmt.Errorf("%s: invalid cursor key", file)
					}
					q.Connection = append(q.Connection, key)
				}
			}
			if strings.HasPrefix(line, "-- @batch ") {
				mapping := strings.Split(strings.TrimSpace(strings.TrimPrefix(line, "-- @batch ")), "=")
				if len(mapping) != 2 || !strings.HasPrefix(mapping[0], "parent.") || !nameRE.MatchString(mapping[1]) {
					return nil, fmt.Errorf("%s: expected -- @batch parent.field=resultColumn", file)
				}
				if q.Batch == nil {
					q.Batch = map[string]string{}
				}
				key := strings.TrimPrefix(mapping[0], "parent.")
				if _, exists := q.Batch[key]; exists {
					return nil, fmt.Errorf("%s: duplicate batch mapping %s", file, key)
				}
				q.Batch[key] = mapping[1]
			}
			if strings.HasPrefix(line, "-- @") && line != "-- @one" && line != "-- @many" && !strings.HasPrefix(line, "-- @batch ") && !strings.HasPrefix(line, "-- @connection ") {
				return nil, fmt.Errorf("%s: unknown annotation %s", file, line)
			}
		}
		q.Source = string(data)
		q.SQL, q.Bindings, err = Bind(string(data))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		q.SQLC, _, err = BindSQLC(string(data))
		if err != nil {
			return nil, err
		}
		bindingNames := map[string]string{}
		for _, b := range q.Bindings {
			name := strings.ToLower(b.Source + "_" + strings.ReplaceAll(b.Path, ".", "_"))
			original := b.Source + "." + b.Path
			if prev, ok := bindingNames[name]; ok && prev != original {
				return nil, fmt.Errorf("%s: sqlc argument name collision between %s and %s", file, prev, original)
			}
			bindingNames[name] = original
		}
		stmt, err := conn.Prepare(ctx, "", q.SQL)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", file, err)
		}
		if len(stmt.Fields) == 0 {
			return nil, fmt.Errorf("%s: query must return columns", file)
		}
		if len(stmt.ParamOIDs) != len(q.Bindings) {
			return nil, fmt.Errorf("%s: parameter count mismatch", file)
		}
		for i, oid := range stmt.ParamOIDs {
			q.Bindings[i].OID = oid
			if q.Bindings[i].Source == "param" {
				q.Bindings[i].Type, err = scalar(oid)
				if err != nil {
					return nil, fmt.Errorf("%s parameter %s: %w", file, q.Bindings[i].Path, err)
				}
			}
		}
		columns := map[string]bool{}
		for _, f := range stmt.Fields {
			if !nameRE.MatchString(f.Name) || strings.HasPrefix(f.Name, "__") || columns[f.Name] {
				return nil, fmt.Errorf("%s: invalid/duplicate column %q; use a quoted GraphQL alias", file, f.Name)
			}
			columns[f.Name] = true
			t, err := scalar(f.DataTypeOID)
			if err != nil {
				return nil, fmt.Errorf("%s column %s: %w", file, f.Name, err)
			}
			q.Columns = append(q.Columns, Column{f.Name, t})
		}
		if err = validateConnection(q); err != nil {
			return nil, err
		}
		paths[path] = q
		queries = append(queries, q)
	}
	for _, q := range queries {
		i := strings.LastIndexByte(q.Path, '.')
		parent := q.Path[:i]
		if parent == "Query" {
			if len(q.Batch) > 0 {
				return nil, fmt.Errorf("%s: root query cannot have batch parent mappings", q.Path)
			}
			for _, b := range q.Bindings {
				if b.Source == "parent" {
					return nil, fmt.Errorf("%s: root has no parent", q.Path)
				}
			}
			continue
		}
		p, ok := paths[parent]
		if !ok {
			return nil, fmt.Errorf("%s: missing parent SQL file %s.sql", q.Path, parent)
		}
		for _, c := range p.Columns {
			if c.Name == q.Field {
				return nil, fmt.Errorf("%s: child conflicts with parent column", q.Path)
			}
		}
		for _, b := range q.Bindings {
			if b.Source == "parent" {
				found := false
				for _, c := range p.Columns {
					if c.Name == b.Path {
						found = true
					}
				}
				if !found {
					return nil, fmt.Errorf("%s: parent column %s not found", q.Path, b.Path)
				}
			}
		}
		mapped := map[string]bool{}
		for _, b := range q.Bindings {
			if b.Source != "parent" {
				continue
			}
			column, ok := q.Batch[b.Path]
			if !ok {
				return nil, fmt.Errorf("%s: batched parent requires -- @batch parent.%s=resultColumn and an array parameter in your SQL (e.g. ANY({{parent.%s}}::int[]))", fileName(q), b.Path, b.Path)
			}
			elem, ok := arrayElement(b.OID)
			if !ok {
				return nil, fmt.Errorf("%s: parent.%s must be an array parameter; write ANY() in the SQL file", q.Path, b.Path)
			}
			typ, err := scalar(elem)
			if err != nil {
				return nil, err
			}
			found := false
			for _, c := range q.Columns {
				if c.Name == column && c.Type == typ {
					found = true
				}
			}
			if !found {
				return nil, fmt.Errorf("%s: batch result column %s missing or has incompatible type", q.Path, column)
			}
			mapped[b.Path] = true
		}
		for key := range q.Batch {
			if !mapped[key] {
				return nil, fmt.Errorf("%s: batch mapping parent.%s has no binding", q.Path, key)
			}
		}
		p.Children = append(p.Children, q)
	}
	return queries, nil
}
func args(q *Query) []Binding {
	var out []Binding
	for _, b := range q.Bindings {
		if b.Source == "param" {
			out = append(out, b)
		}
	}
	if len(q.Connection) > 0 {
		out = append(out, Binding{Source: "param", Path: "first", Type: "Int", Nullable: true}, Binding{Source: "param", Path: "after", Type: "String", Nullable: true}, Binding{Source: "param", Path: "last", Type: "Int", Nullable: true}, Binding{Source: "param", Path: "before", Type: "String", Nullable: true})
	}
	return out
}
func field(q *Query) string {
	s := q.Field
	a := args(q)
	if len(a) > 0 {
		var parts []string
		for _, b := range a {
			typ := b.Type
			if !b.Nullable {
				typ += "!"
			}
			parts = append(parts, b.Path+": "+typ)
		}
		s += "(" + strings.Join(parts, ", ") + ")"
	}
	s += ": "
	if len(q.Connection) > 0 {
		return s + q.Type + "Connection!"
	}
	if q.One {
		return s + q.Type
	}
	return s + "[" + q.Type + "!]!"
}
func Render(queries []*Query) (string, string, map[string]string, error) {
	var schema, sql strings.Builder
	schema.WriteString("# Code generated by sqlgql. DO NOT EDIT.\ntype Query {\n")
	for _, q := range queries {
		if strings.Count(q.Path, ".") == 1 {
			fmt.Fprintf(&schema, "  %s\n", field(q))
		}
	}
	schema.WriteString("}\n")
	connected := false
	for _, q := range queries {
		if len(q.Connection) > 0 {
			connected = true
		}
	}
	if connected {
		schema.WriteString("\ntype PageInfo { hasNextPage: Boolean! hasPreviousPage: Boolean! startCursor: String endCursor: String }\n")
	}
	for _, q := range queries {
		if len(q.Connection) > 0 {
			fmt.Fprintf(&schema, "\ntype %sConnection { edges: [%sEdge!]! pageInfo: PageInfo! }\ntype %sEdge { cursor: String! node: %s! }\n", q.Type, q.Type, q.Type, q.Type)
		}
	}
	used := map[string]bool{}
	for _, q := range queries {
		for _, c := range q.Columns {
			used[strings.Trim(c.Type, "[]!")] = true
		}
		for _, b := range q.Bindings {
			used[strings.Trim(b.Type, "[]!")] = true
		}
	}
	for _, name := range []string{"UUID", "JSON", "BigInt"} {
		if used[name] {
			fmt.Fprintf(&schema, "\nscalar %s\n", name)
		}
	}
	for _, q := range queries {
		fmt.Fprintf(&schema, "\ntype %s {\n", q.Type)
		for _, c := range q.Columns {
			fmt.Fprintf(&schema, "  %s: %s\n", c.Name, c.Type)
		}
		for _, child := range q.Children {
			fmt.Fprintf(&schema, "  %s\n", field(child))
		}
		schema.WriteString("}\n")
	}
	sql.WriteString("-- Code generated by sqlgql. DO NOT EDIT.\n")
	for _, q := range queries {
		cardinality := "many"
		if q.One && len(q.Batch) == 0 {
			cardinality = "one"
		}
		fmt.Fprintf(&sql, "\n-- name: %s :%s\n%s;\n", q.Func, cardinality, strings.TrimSuffix(strings.TrimSpace(q.SQLC), ";"))
	}
	operations := map[string]string{}
	for _, q := range queries {
		if strings.Count(q.Path, ".") != 1 {
			continue
		}
		var vars []string
		selection := selectQuery(q, 1, &vars)
		header := "query " + q.Type
		if len(vars) > 0 {
			header += "(" + strings.Join(vars, ", ") + ")"
		}
		operations[q.Type+".graphql"] = header + " {\n" + selection + "}\n"
	}
	s, err := gqlparser.LoadSchema(&ast.Source{Name: "generated", Input: schema.String()})
	if err != nil {
		return "", "", nil, err
	}
	for name, op := range operations {
		if _, errs := gqlparser.LoadQuery(s, op); len(errs) > 0 {
			return "", "", nil, fmt.Errorf("%s: %s", name, errs.Error())
		}
	}
	return schema.String(), sql.String(), operations, nil
}
func selectQuery(q *Query, depth int, vars *[]string) string {
	indent := strings.Repeat("  ", depth)
	s := indent + q.Field
	var a []string
	for _, b := range args(q) {
		v := strings.ReplaceAll(q.Path, ".", "_") + "_" + b.Path
		typ := b.Type
		if !b.Nullable {
			typ += "!"
		}
		*vars = append(*vars, "$"+v+": "+typ)
		a = append(a, b.Path+": $"+v)
	}
	if len(a) > 0 {
		s += "(" + strings.Join(a, ", ") + ")"
	}
	s += " {\n" + indent + "  __typename\n"
	outer := indent
	if len(q.Connection) > 0 {
		s += indent + "  pageInfo { __typename hasNextPage hasPreviousPage startCursor endCursor }\n" + indent + "  edges {\n" + indent + "    __typename\n" + indent + "    cursor\n" + indent + "    node {\n" + indent + "      __typename\n"
		depth += 2
		indent = strings.Repeat("  ", depth)
	}
	for _, c := range q.Columns {
		s += indent + "  " + c.Name + "\n"
	}
	for _, c := range q.Children {
		s += selectQuery(c, depth+1, vars)
	}
	if len(q.Connection) > 0 {
		s += indent + "}\n" + outer + "  }\n"
		indent = outer
	}
	return s + indent + "}\n"
}
func Generate(ctx context.Context, conn *pgx.Conn, dir, out string) error {
	qs, err := Inspect(ctx, conn, dir)
	if err != nil {
		return err
	}
	schema, sql, ops, err := Render(qs)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Join(out, "operations"), 0755); err != nil {
		return err
	}
	manifest, err := json.MarshalIndent(qs, "", "  ")
	if err != nil {
		return err
	}
	files := map[string]string{"schema.graphqls": schema, "queries.sql": sql, "manifest.json": string(manifest) + "\n"}
	for n, v := range ops {
		files[filepath.Join("operations", n)] = v
	}
	var names []string
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		if err = os.WriteFile(filepath.Join(out, n), []byte(files[n]), 0644); err != nil {
			return err
		}
	}
	// This directory is generator-owned; remove operations for deleted SQL roots.
	old, _ := filepath.Glob(filepath.Join(out, "operations", "*.graphql"))
	for _, p := range old {
		if _, ok := ops[filepath.Base(p)]; !ok {
			if err = os.Remove(p); err != nil {
				return err
			}
		}
	}
	return nil
}

func fileName(q *Query) string { return q.Path + ".sql" }
func arrayElement(oid uint32) (uint32, bool) {
	v, ok := map[uint32]uint32{1000: 16, 1016: 20, 1005: 21, 1007: 23, 1021: 700, 1022: 701, 1009: 25, 1014: 1042, 1015: 1043, 2951: 2950, 199: 114, 3807: 3802}[oid]
	return v, ok
}

func validateConnection(q *Query) error {
	seen := map[string]bool{}
	for _, key := range q.Connection {
		if seen[key] {
			return fmt.Errorf("%s: duplicate cursor key %s", q.Path, key)
		}
		seen[key] = true
		found := false
		for _, c := range q.Columns {
			if c.Name == key && c.Type != "JSON" && !strings.HasPrefix(c.Type, "[") {
				found = true
			}
		}
		if !found {
			return fmt.Errorf("%s: cursor key %s must be a scalar result column", q.Path, key)
		}
	}
	if len(q.Connection) > 0 && q.One {
		return fmt.Errorf("%s: @connection requires @many", q.Path)
	}
	required := map[string]bool{"limit": false, "backward": false}
	for _, key := range q.Connection {
		required["after."+key] = false
		required["before."+key] = false
	}
	for _, b := range q.Bindings {
		if b.Source == "page" {
			if len(q.Connection) == 0 {
				return fmt.Errorf("%s: page bindings require @connection", q.Path)
			}
			if _, ok := required[b.Path]; !ok {
				return fmt.Errorf("%s: unknown page binding %s", q.Path, b.Path)
			}
			required[b.Path] = true
			if b.Path == "limit" && b.OID != 23 {
				return fmt.Errorf("page.limit must be PostgreSQL int4")
			}
			if b.Path == "backward" && b.OID != 16 {
				return fmt.Errorf("page.backward must be PostgreSQL bool")
			}
		}
		if len(q.Connection) > 0 && b.Source == "param" {
			switch b.Path {
			case "first", "after", "last", "before":
				return fmt.Errorf("%s: %s is reserved for Relay pagination", q.Path, b.Path)
			}
		}
	}
	if len(q.Connection) > 0 {
		var missing []string
		for name, found := range required {
			if !found {
				missing = append(missing, name)
			}
		}
		sort.Strings(missing)
		if len(missing) > 0 {
			return fmt.Errorf("%s: connection SQL must bind page.%s", q.Path, strings.Join(missing, ", page."))
		}
	}
	return nil
}
