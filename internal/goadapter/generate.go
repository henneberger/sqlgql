// Package goadapter is an optional gqlgen/sqlc integration, separate from the SQL compiler.
package goadapter

import (
	"encoding/json"
	"fmt"
	"go/types"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/henneberger/sqlgql/internal/compiler"
	"github.com/99designs/gqlgen/api"
	"github.com/99designs/gqlgen/codegen"
	"github.com/99designs/gqlgen/codegen/config"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/imports"
)

func Generate(manifest, configFile, sqlcPackage string) error {
	raw, err := os.ReadFile(manifest)
	if err != nil {
		return err
	}
	var queries []*compiler.Query
	if err = json.Unmarshal(raw, &queries); err != nil {
		return err
	}
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		return err
	}
	// Supply scalar defaults without requiring per-query gqlgen configuration.
	for _, name := range []string{"UUID", "JSON", "BigInt"} {
		if len(cfg.Models[name].Model) == 0 {
			cfg.Models[name] = config.TypeMapEntry{Model: config.StringList{"github.com/henneberger/sqlgql/binding." + name}}
		}
	}
	paths := map[string]*compiler.Query{}
	for _, q := range queries {
		paths[q.Path] = q
	}
	for _, q := range queries {
		parent := q.Path[:strings.LastIndex(q.Path, ".")]
		if parent == "Query" {
			continue
		}
		p := paths[parent]
		if p == nil {
			return fmt.Errorf("missing parent %s", parent)
		}
		entry := cfg.Models[p.Type]
		if entry.Fields == nil {
			entry.Fields = map[string]config.TypeMapField{}
		}
		field := entry.Fields[q.Field]
		field.Resolver = true
		entry.Fields[q.Field] = field
		cfg.Models[p.Type] = entry
	}
	pkgs, err := packages.Load(&packages.Config{Mode: packages.NeedTypes | packages.NeedImports | packages.NeedDeps}, sqlcPackage)
	if err != nil {
		return err
	}
	if len(pkgs) != 1 || len(pkgs[0].Errors) > 0 {
		return fmt.Errorf("load sqlc package: %v", pkgs)
	}
	plugin := &generator{queries: queries, paths: paths, db: pkgs[0].Types, imports: map[string]string{sqlcPackage: "db"}}
	return api.Generate(cfg, api.ReplacePlugin(plugin))
}

type generator struct {
	queries []*compiler.Query
	paths   map[string]*compiler.Query
	db      *types.Package
	imports map[string]string
	local   string
	objects map[string]*codegen.Object
	body    strings.Builder
}

func (g *generator) Name() string { return "resolvergen" }
func (g *generator) typ(t types.Type) string {
	return types.TypeString(t, func(p *types.Package) string {
		if p.Path() == g.local {
			return ""
		}
		if alias, ok := g.imports[p.Path()]; ok {
			return alias
		}
		alias := p.Name()
		for _, used := range g.imports {
			if used == alias {
				alias = fmt.Sprintf("pkg%d", len(g.imports))
				break
			}
		}
		g.imports[p.Path()] = alias
		return alias
	})
}
func (g *generator) write(format string, args ...any) { fmt.Fprintf(&g.body, format, args...) }
func (g *generator) method(name string) (*types.Signature, error) {
	obj := g.db.Scope().Lookup("Queries")
	if obj == nil {
		return nil, fmt.Errorf("sqlc Queries type missing")
	}
	method, _, _ := types.LookupFieldOrMethod(types.NewPointer(obj.Type()), true, g.db, name)
	if method == nil {
		return nil, fmt.Errorf("sqlc method %s missing; run sqlc first", name)
	}
	return method.Type().(*types.Signature), nil
}
func fields(t types.Type) map[string]*types.Var {
	if p, ok := t.(*types.Pointer); ok {
		t = p.Elem()
	}
	s, ok := t.Underlying().(*types.Struct)
	if !ok {
		return nil
	}
	out := map[string]*types.Var{}
	for i := 0; i < s.NumFields(); i++ {
		tag := reflect.StructTag(s.Tag(i))
		name := tag.Get("db")
		if name == "" {
			name = strings.Split(tag.Get("json"), ",")[0]
		}
		if name != "" {
			out[name] = s.Field(i)
		}
	}
	return out
}
func objectField(o *codegen.Object, name string) *codegen.Field {
	for _, f := range o.Fields {
		if f.Name == name {
			return f
		}
	}
	return nil
}
func (g *generator) parent(q *compiler.Query) string {
	path := q.Path[:strings.LastIndex(q.Path, ".")]
	if path == "Query" {
		return path
	}
	return g.paths[path].Type
}
func hasParents(q *compiler.Query) bool {
	for _, b := range q.Bindings {
		if b.Source == "parent" {
			return true
		}
	}
	return false
}
func bindName(b compiler.Binding) string {
	return b.Source + "_" + strings.ToLower(strings.ReplaceAll(b.Path, ".", "_"))
}
func (g *generator) signature(q *compiler.Query) (string, string, string) {
	f := objectField(g.objects[g.parent(q)], q.Field)
	decl := "ctx context.Context"
	call := "ctx"
	if g.parent(q) != "Query" {
		decl += ", obj " + g.typ(g.objects[g.parent(q)].Reference())
		call += ", obj"
	}
	for _, a := range f.Args {
		decl += ", " + a.VarName + " " + g.typ(a.TypeReference.GO)
		call += ", " + a.VarName
	}
	return decl, call, g.typ(f.TypeReference.GO)
}
func (g *generator) GenerateCode(data *codegen.Data) error {
	g.local = data.Config.Exec.ImportPath()
	g.objects = map[string]*codegen.Object{}
	for _, o := range data.Objects {
		g.objects[o.Name] = o
	}
	for _, q := range g.queries {
		if g.objects[q.Type] == nil {
			return fmt.Errorf("GraphQL type %s missing", q.Type)
		}
	}
	g.write(`// Code generated by sqlgql-go. DO NOT EDIT.
package %s
`, data.Config.Resolver.Package)
	// Imports are inserted after generation, using actual sqlc/gqlgen Go types.
	g.write(`
type SQLResolvers struct {
 Queries db.Querier
 Authorizer binding.Authorizer
 Environment func(context.Context) map[string]any
 Overrides Overrides
 BatchWait time.Duration
 BatchSize int
}
type Overrides struct {
`)
	for _, q := range g.queries {
		decl, _, ret := g.signature(q)
		g.write("%s func(%s) (%s,error)\n", q.Func, decl, ret)
	}
	g.write("}\ntype loaderContextKey struct{}\ntype requestLoaders struct {\nenv map[string]any\n")
	for _, q := range g.queries {
		if g.parent(q) != "Query" {
			ret := g.dataResult(q)
			g.write("%s *dataloader.Loader[string,%s]\n", q.Func, ret)
		}
	}
	g.write("}\n// WithLoaders attaches operation-local caches; call from your gqlgen operation middleware.\nfunc(r *SQLResolvers) WithLoaders(ctx context.Context) context.Context {\nstate:=&requestLoaders{}\nwait:=r.BatchWait;if wait<=0{wait=2*time.Millisecond};size:=r.BatchSize;if size<=0{size=128}\n_ = wait; _ = size\nif r.Environment!=nil{state.env=r.Environment(ctx)}\n")
	for _, q := range g.queries {
		if g.parent(q) != "Query" {
			ret := g.dataResult(q)
			g.write("state.%s=dataloader.NewBatchedLoader(r.batch%s,dataloader.WithWait[string,%s](wait),dataloader.WithBatchCapacity[string,%s](size))\n", q.Func, q.Func, ret, ret)
		}
	}
	g.write("return context.WithValue(ctx,loaderContextKey{},state)\n}\n")
	parents := map[string]bool{}
	for _, q := range g.queries {
		parents[g.parent(q)] = true
	}
	var names []string
	for p := range parents {
		names = append(names, p)
	}
	sort.Strings(names)
	for _, p := range names {
		g.write("func(r *SQLResolvers) %s() %sResolver{return &sql%sResolver{r}}\ntype sql%sResolver struct{*SQLResolvers}\n", p, p, p, p)
	}
	for _, q := range g.queries {
		if len(q.Connection) > 0 {
			g.connection(q)
		}
		if err := g.resolver(q); err != nil {
			return err
		}
		if g.parent(q) != "Query" {
			if err := g.batch(q); err != nil {
				return err
			}
		}
	}
	baseImports := map[string]string{"context": "context", "encoding/json": "json", "errors": "errors", "fmt": "fmt", "time": "time", "github.com/jackc/pgx/v5": "pgx", "github.com/graph-gophers/dataloader/v7": "dataloader", "github.com/henneberger/sqlgql/binding": "binding"}
	for p, a := range baseImports {
		g.imports[p] = a
	}
	var imp strings.Builder
	imp.WriteString("import (\n")
	var paths []string
	for p := range g.imports {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	for _, p := range paths {
		fmt.Fprintf(&imp, "%s %q\n", g.imports[p], p)
	}
	imp.WriteString(")\n")
	source := g.body.String()
	at := strings.Index(source, "\ntype SQLResolvers")
	source = source[:at] + "\n" + imp.String() + source[at:]
	dest := filepath.Join(data.Config.Resolver.Dir(), "sqlgql_gen.go")
	formatted, err := imports.Process(dest, []byte(source), nil)
	if err != nil {
		return fmt.Errorf("format adapters: %w\n%s", err, source)
	}
	return os.WriteFile(dest, formatted, 0644)
}

func (g *generator) resolver(q *compiler.Query) error {
	parent := g.parent(q)
	f := objectField(g.objects[parent], q.Field)
	decl, call, ret := g.signature(q)
	method, err := g.method(q.Func)
	if err != nil {
		return err
	}
	g.write("func(r *sql%sResolver) %s(%s)(%s,error){\n", parent, f.GoFieldName, decl, ret)
	g.write("state,ok:=ctx.Value(loaderContextKey{}).(*requestLoaders);if !ok{return nil,fmt.Errorf(\"operation loaders not installed: call WithLoaders\")}\nif r.Authorizer==nil{return nil,binding.ErrForbidden}\n")
	g.write("params:=map[string]any{")
	for _, a := range f.Args {
		g.write("%q:%s,", a.Name, a.VarName)
	}
	g.write("}\nparent:=map[string]any{\n")
	if parent != "Query" {
		for _, b := range q.Bindings {
			if b.Source == "parent" {
				pf := objectField(g.objects[parent], b.Path)
				if pf == nil {
					return fmt.Errorf("parent field %s missing", b.Path)
				}
				g.write("%q:obj.%s,\n", b.Path, pf.GoFieldName)
			}
		}
	}
	g.write("}\nif err:=r.Authorizer.Authorize(ctx,binding.Decision{Path:%q,Env:state.env,Parent:parent,Params:params});err!=nil{return nil,err}\n", q.Path)
	g.write("if r.Overrides.%s!=nil{return r.Overrides.%s(%s)}\n", q.Func, q.Func, call)
	g.write("if r.Queries==nil{return nil,fmt.Errorf(\"sqlc queries are required\")}\n")
	if len(q.Connection) > 0 {
		g.write("page,err:=binding.PreparePage(%q,[]string{", q.Path)
		for _, key := range q.Connection {
			g.write("%q,", key)
		}
		g.write("},state.env,parent,params,first,after,last,before);if err!=nil{return nil,err}\n")
	}
	queryArgs := "ctx"
	if method.Params().Len() > 1 {
		paramType := method.Params().At(1).Type()
		paramFields := fields(paramType)
		g.write("var input %s\n", g.typ(paramType))
		for i, b := range q.Bindings {
			v := paramFields[bindName(b)]
			if v == nil {
				return fmt.Errorf("%s: sqlc parameter %s missing (enable emit_db_tags and query_parameter_limit: 0)", q.Path, bindName(b))
			}
			expr := ""
			switch b.Source {
			case "param":
				expr = "params[" + fmt.Sprintf("%q", b.Path) + "]"
			case "parent":
				expr = "parent[" + fmt.Sprintf("%q", b.Path) + "]"
			case "page":
				expr = "page.Value(" + fmt.Sprintf("%q", b.Path) + ")"
			case "env":
				g.write("value%d,err:=binding.Lookup(state.env,%q);if err!=nil{return nil,err}\n", i, b.Path)
				expr = fmt.Sprintf("value%d", i)
			}
			convert := "Convert"
			if b.OID == 20 || b.OID == 1016 {
				convert = "ConvertBigInt"
			}
			if b.OID == 700 || b.OID == 1021 {
				convert = "ConvertFloat32"
			}
			if b.Source == "page" && (strings.HasPrefix(b.Path, "after.") || strings.HasPrefix(b.Path, "before.")) {
				convert = "Copy"
				if b.OID == 20 {
					convert = "OptionalBigInt"
				}
			}
			if b.Source == "parent" {
				slice, ok := v.Type().(*types.Slice)
				if !ok {
					return fmt.Errorf("%s: sqlc parent parameter must be an array", q.Path)
				}
				g.write("bound%d,err:=binding.%s[%s](%s);if err!=nil{return nil,fmt.Errorf(%q,err)}\ninput.%s=%s{bound%d}\n", i, convert, g.typ(slice.Elem()), expr, q.Path+" parent."+b.Path+": %w", v.Name(), g.typ(v.Type()), i)
			} else {
				g.write("bound%d,err:=binding.%s[%s](%s);if err!=nil{return nil,fmt.Errorf(%q,err)}\ninput.%s=bound%d\n", i, convert, g.typ(v.Type()), expr, q.Path+" "+b.Source+"."+b.Path+": %w", v.Name(), i)
			}
		}
		queryArgs += ",input"
	}
	if g.parent(q) != "Query" {
		if method.Params().Len() == 1 {
			g.write("input:=struct{}{}\n")
		}
		if len(q.Connection) > 0 {
			g.write("key,err:=json.Marshal(input);if err!=nil{return nil,err};rows,err:=state.%s.Load(ctx,string(key))();if err!=nil{return nil,err};return connect%s(rows,page)\n}\n", q.Func, q.Func)
		} else {
			g.write("key,err:=json.Marshal(input);if err!=nil{return nil,err};return state.%s.Load(ctx,string(key))()\n}\n", q.Func)
		}
		return nil
	}
	if q.One {
		g.write("row,err:=r.Queries.%s(%s);if errors.Is(err,pgx.ErrNoRows){return nil,nil};if err!=nil{return nil,err}\n", q.Func, queryArgs)
		if err := g.row(q, method.Results().At(0).Type(), "row", "result"); err != nil {
			return err
		}
		g.write("return result,nil\n}\n")
	} else {
		g.write("rows,err:=r.Queries.%s(%s);if err!=nil{return nil,err}\nresult:=make(%s,0,len(rows));for _,row:=range rows{\n", q.Func, queryArgs, g.dataResult(q))
		slice, ok := method.Results().At(0).Type().(*types.Slice)
		if !ok {
			return fmt.Errorf("%s: expected sqlc slice result", q.Path)
		}
		if err := g.row(q, slice.Elem(), "row", "item"); err != nil {
			return err
		}
		if len(q.Connection) > 0 {
			g.write("result=append(result,item)\n};return connect%s(result,page)\n}\n", q.Func)
		} else {
			g.write("result=append(result,item)\n};return result,nil\n}\n")
		}
	}
	return nil
}

// Emit explicit field assignments using the names selected by sqlc and gqlgen.
func (g *generator) row(q *compiler.Query, rowType types.Type, src, dest string) error {
	rowFields := fields(rowType)
	target := g.objects[q.Type]
	g.write("%s:=new(%s)\n", dest, g.typ(target.Type))
	for i, c := range q.Columns {
		v := rowFields[c.Name]
		f := objectField(target, c.Name)
		if v == nil || f == nil {
			return fmt.Errorf("%s: cannot map column %s", q.Path, c.Name)
		}
		dstType := f.TypeReference.GO
		if types.Identical(v.Type(), dstType) {
			g.write("%s.%s=%s.%s\n", dest, f.GoFieldName, src, v.Name())
			continue
		}
		if ptr, ok := dstType.(*types.Pointer); ok && safeScalar(v.Type(), ptr.Elem()) {
			g.write("field%d:=%s(%s.%s);%s.%s=&field%d\n", i, g.typ(ptr.Elem()), src, v.Name(), dest, f.GoFieldName, i)
			continue
		}
		if from, ok := v.Type().(*types.Pointer); ok {
			if to, ok := dstType.(*types.Pointer); ok && safeScalar(from.Elem(), to.Elem()) {
				g.write("if %s.%s!=nil{field%d:=%s(*%s.%s);%s.%s=&field%d}\n", src, v.Name(), i, g.typ(to.Elem()), src, v.Name(), dest, f.GoFieldName, i)
				continue
			}
		}
		if from, ok := v.Type().(*types.Slice); ok {
			if to, ok := dstType.(*types.Slice); ok {
				fromElem, fromOK := from.Elem().(*types.Pointer)
				toElem, toOK := to.Elem().(*types.Pointer)
				if fromOK && toOK && safeScalar(fromElem.Elem(), toElem.Elem()) {
					g.write("if %s.%s!=nil{%s.%s=make(%s,len(%s.%s));for i,value:=range %s.%s{if value!=nil{element:=%s(*value);%s.%s[i]=&element}}}\n", src, v.Name(), dest, f.GoFieldName, g.typ(dstType), src, v.Name(), src, v.Name(), g.typ(toElem.Elem()), dest, f.GoFieldName)
					continue
				}
			}
		}
		g.write("field%d,err:=binding.Copy[%s](%s.%s);if err!=nil{return nil,err};%s.%s=field%d\n", i, g.typ(dstType), src, v.Name(), dest, f.GoFieldName, i)
	}
	return nil
}

func (g *generator) batch(q *compiler.Query) error {
	if !hasParents(q) {
		return g.uncorrelatedBatch(q)
	}
	method, err := g.method(q.Func)
	if err != nil {
		return err
	}
	bm, err := g.method(q.Func)
	if err != nil {
		return err
	}
	paramType := method.Params().At(1).Type()
	batchType := bm.Params().At(1).Type()
	pf, bf := fields(paramType), fields(batchType)
	ret := g.dataResult(q)
	rowType := bm.Results().At(0).Type().(*types.Slice).Elem()
	g.write("func(r *SQLResolvers) batch%s(ctx context.Context,keys []string) []*dataloader.Result[%s]{\n", q.Func, ret)
	g.write("results:=make([]*dataloader.Result[%s],len(keys));inputs:=make([]%s,len(keys));groups:=map[string][]int{}\n", ret, g.typ(paramType))
	g.write("for i,key:=range keys{results[i]=&dataloader.Result[%s]{}\n", ret)
	if !q.One {
		g.write("results[i].Data=make(%s,0)\n", ret)
	}
	g.write("if err:=json.Unmarshal([]byte(key),&inputs[i]);err!=nil{results[i].Error=err;continue};group:=inputs[i];var zero %s\n", g.typ(paramType))
	for _, b := range q.Bindings {
		if b.Source == "parent" {
			v := pf[bindName(b)]
			g.write("group.%s=zero.%s\n", v.Name(), v.Name())
		}
	}
	for _, b := range q.Bindings {
		if b.Source == "parent" {
			v := pf[bindName(b)]
			g.write("if len(inputs[i].%s)!=1{results[i].Error=fmt.Errorf(\"invalid loader parent key\");continue}\n", v.Name())
		}
	}
	g.write("encoded,err:=json.Marshal(group);if err!=nil{results[i].Error=err;continue};groups[string(encoded)]=append(groups[string(encoded)],i)\n}\n")
	g.write("for _,positions:=range groups{\nfetch:=func()error{\nvar batch %s\n", g.typ(batchType))
	for _, b := range q.Bindings {
		v, w := pf[bindName(b)], bf[bindName(b)]
		if v == nil || w == nil {
			return fmt.Errorf("%s: batch parameter missing", q.Path)
		}
		if b.Source == "parent" {
			g.write("for _,i:=range positions{batch.%s=append(batch.%s,inputs[i].%s...)}\n", w.Name(), w.Name(), v.Name())
		} else {
			g.write("batch.%s=inputs[positions[0]].%s\n", w.Name(), v.Name())
		}
	}
	// Index each input tuple. The SQL file returns the declared parent key columns.
	g.write("routing:=map[string][]int{}\nfor _,i:=range positions{tuple:=[]any{\n")
	for _, b := range q.Bindings {
		if b.Source == "parent" {
			v := pf[bindName(b)]
			g.write("inputs[i].%s[0],\n", v.Name())
		}
	}
	g.write("};key,err:=json.Marshal(tuple);if err!=nil{return err};routing[string(key)]=append(routing[string(key)],i)}\n")
	g.write("rows,err:=r.Queries.%s(ctx,batch);if err!=nil{return err}\nfor _,row:=range rows{tuple:=[]any{\n", q.Func)
	for _, b := range q.Bindings {
		if b.Source == "parent" {
			field := fields(rowType)[q.Batch[b.Path]]
			if field == nil {
				return fmt.Errorf("%s: batch routing column %s missing", q.Path, q.Batch[b.Path])
			}
			g.write("row.%s,\n", field.Name())
		}
	}
	g.write("};key,err:=json.Marshal(tuple);if err!=nil{return err};targets,ok:=routing[string(key)];if !ok{return fmt.Errorf(\"SQL returned an unrequested parent key\")}\n")
	// Mapping uses the same generated mapper as direct queries, wrapped for errors.
	g.write("mapRow:=func()(*%s,error){\n", g.typ(g.objects[q.Type].Type))
	if err := g.row(q, rowType, "row", "item"); err != nil {
		return err
	}
	g.write("return item,nil\n}\nitem,err:=mapRow();if err!=nil{return err}\n")
	if q.One {
		g.write("for _,position:=range targets{if results[position].Data!=nil{return fmt.Errorf(\"@one query returned multiple rows for a parent\")};results[position].Data=item}\n")
	} else {
		g.write("for _,position:=range targets{results[position].Data=append(results[position].Data,item)}\n")
	}
	g.write("};return nil\n}\nif err:=fetch();err!=nil{for _,i:=range positions{results[i].Error=err}}\n}\nreturn results\n}\n")
	return nil
}

func safeScalar(src, dst types.Type) bool {
	if types.Identical(src, dst) {
		return true
	}
	a, aok := src.(*types.Basic)
	b, bok := dst.(*types.Basic)
	if !aok || !bok {
		return false
	}
	return (b.Kind() == types.Int && (a.Kind() == types.Int16 || a.Kind() == types.Int32)) || (b.Kind() == types.Float64 && a.Kind() == types.Float32)
}

func (g *generator) uncorrelatedBatch(q *compiler.Query) error {
	method, err := g.method(q.Func)
	if err != nil {
		return err
	}
	ret := g.dataResult(q)
	g.write("func(r *SQLResolvers) batch%s(ctx context.Context,keys []string) []*dataloader.Result[%s]{\nresults:=make([]*dataloader.Result[%s],len(keys));for i,key:=range keys{\nfetch:=func()(%s,error){\n", q.Func, ret, ret, ret)
	call := "ctx"
	if method.Params().Len() > 1 {
		g.write("var input %s;if err:=json.Unmarshal([]byte(key),&input);err!=nil{return nil,err}\n", g.typ(method.Params().At(1).Type()))
		call += ",input"
	} else {
		g.write("_ = key\n")
	}
	if q.One {
		g.write("row,err:=r.Queries.%s(%s);if errors.Is(err,pgx.ErrNoRows){return nil,nil};if err!=nil{return nil,err}\n", q.Func, call)
		if err := g.row(q, method.Results().At(0).Type(), "row", "result"); err != nil {
			return err
		}
		g.write("return result,nil\n")
	} else {
		g.write("rows,err:=r.Queries.%s(%s);if err!=nil{return nil,err};result:=make(%s,0,len(rows));for _,row:=range rows{\n", q.Func, call, ret)
		if err := g.row(q, method.Results().At(0).Type().(*types.Slice).Elem(), "row", "item"); err != nil {
			return err
		}
		g.write("result=append(result,item)\n};return result,nil\n")
	}
	g.write("};value,err:=fetch();results[i]=&dataloader.Result[%s]{Data:value,Error:err}\n};return results\n}\n", ret)
	return nil
}

func (g *generator) dataResult(q *compiler.Query) string {
	if len(q.Connection) > 0 {
		return "[]*" + g.typ(g.objects[q.Type].Type)
	}
	_, _, ret := g.signature(q)
	return ret
}
func (g *generator) connection(q *compiler.Query) {
	node := g.typ(g.objects[q.Type].Type)
	conn := g.typ(g.objects[q.Type+"Connection"].Type)
	edge := g.typ(g.objects[q.Type+"Edge"].Type)
	info := g.typ(g.objects["PageInfo"].Type)
	g.write("func connect%s(rows []*%s,page *binding.Page)(*%s,error){\nnodes,more:=binding.PageRows(rows,page);result:=&%s{Edges:make([]*%s,0,len(nodes)),PageInfo:&%s{}}\nif page.Backward{result.PageInfo.HasPreviousPage=more}else{result.PageInfo.HasNextPage=more}\nfor _,node:=range nodes{cursor,err:=page.Cursor(", q.Func, node, conn, conn, edge, info)
	for _, key := range q.Connection {
		g.write("node.%s,", objectField(g.objects[q.Type], key).GoFieldName)
	}
	g.write(");if err!=nil{return nil,err};result.Edges=append(result.Edges,&%s{Cursor:cursor,Node:node})}\nif len(result.Edges)>0{start:=result.Edges[0].Cursor;end:=result.Edges[len(result.Edges)-1].Cursor;result.PageInfo.StartCursor=&start;result.PageInfo.EndCursor=&end};return result,nil\n}\n", edge)
}
