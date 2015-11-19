/*
Copyright 2015 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"k8s.io/kubernetes/cmd/libs/go2idl/generator"
	"k8s.io/kubernetes/cmd/libs/go2idl/namer"
	"k8s.io/kubernetes/cmd/libs/go2idl/types"
)

// genProtoIDL produces a .proto IDL.
type genProtoIDL struct {
	generator.DefaultGen
	localPackage   types.Name
	localGoPackage types.Name
	imports        *ImportTracker

	generateAll bool
}

func (g *genProtoIDL) PackageVars(c *generator.Context) []string {
	return []string{
		"option (gogoproto.marshaler_all) = true;",
		"option (gogoproto.sizer_all) = true;",
		"option (gogoproto.unmarshaler_all) = true;",
		"option (gogoproto.goproto_unrecognized_all) = false;",
		"option (gogoproto.goproto_stringer_all) = false;",
		"option (gogoproto.goproto_enum_prefix_all) = false;",
		"option (gogoproto.goproto_getters_all) = false;",
		fmt.Sprintf("option go_package = %q;", g.localGoPackage.Name),
	}
}
func (g *genProtoIDL) Filename() string { return g.OptionalName + ".proto" }
func (g *genProtoIDL) FileType() string { return "protoidl" }
func (g *genProtoIDL) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		// The local namer returns the correct protobuf name for a proto type
		// in the context of a package
		"local": localNamer{g.localPackage},
	}
}

// Filter ignores types that are identified as not exportable.
func (g *genProtoIDL) Filter(c *generator.Context, t *types.Type) bool {
	flags := types.ExtractCommentTags("+", t.CommentLines)
	switch {
	case flags["genprotoidl"] == "false":
		return false
	case flags["genprotoidl"] == "true":
		return true
	case !g.generateAll:
		return false
	}
	seen := map[*types.Type]bool{}
	return isProtoable(seen, t)
}

func isProtoable(seen map[*types.Type]bool, t *types.Type) bool {
	if seen[t] {
		// be optimistic in the case of type cycles.
		return true
	}
	seen[t] = true
	switch t.Kind {
	case types.Builtin:
		return true
	case types.Alias:
		return isProtoable(seen, t.Underlying)
	case types.Slice, types.Pointer:
		return isProtoable(seen, t.Elem)
	case types.Map:
		return isProtoable(seen, t.Key) && isProtoable(seen, t.Elem)
	case types.Struct:
		for _, m := range t.Members {
			if !isProtoable(seen, m.Type) {
				return false
			}
		}
		return true
	case types.Func, types.Chan:
		return false
	default:
		// Uncomment this if types aren't showing up
		log.Printf("type is not protable: %s", t.Name)
		return false
	}
}

func (g *genProtoIDL) Imports(c *generator.Context) (imports []string) {
	return g.imports.ImportLines()
}

// GenerateType makes the body of a file implementing a set for type t.
func (g *genProtoIDL) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	b := bodyGen{
		locator: &protobufLocator{
			namer:   c.Namers["proto"].(ProtobufFromGoNamer),
			tracker: g.imports,

			localGoPackage: g.localGoPackage.Package,
		},
		localPackage: g.localPackage,

		t: t,
	}
	switch t.Kind {
	case types.Struct:
		b.doStruct(sw)
	default:
		b.unknown(sw)
	}
	return sw.Error()
}

// ProtobufFromGoNamer finds the protobuf name of a type (and its package, and
// the package path) from its Go name.
type ProtobufFromGoNamer interface {
	GoNameToProtoName(name types.Name) types.Name
}

type ProtobufLocator interface {
	ProtoTypeFor(t *types.Type) (*types.Type, error)
	CastTypeName(name types.Name) string
}

type protobufLocator struct {
	namer   ProtobufFromGoNamer
	tracker namer.ImportTracker

	localGoPackage string
}

// CastTypeName returns the cast type name of a Go type
// TODO: delegate to a new localgo namer?
func (p protobufLocator) CastTypeName(name types.Name) string {
	if name.Package == p.localGoPackage {
		return name.Name
	}
	return name.String()
}

// ProtoTypeFor locates a Protobuf type for the provided Go type (if possible).
func (p protobufLocator) ProtoTypeFor(t *types.Type) (*types.Type, error) {
	switch {
	// we've already converted the type, or it's a map
	case t.Kind == typesKindProtobuf || t.Kind == types.Map:
		p.tracker.AddType(t)
		return t, nil
	}
	// it's a fundamental type
	if t, ok := isFundamentalProtoType(t); ok {
		p.tracker.AddType(t)
		return t, nil
	}
	// it's a message
	if t.Kind == types.Struct {
		t := &types.Type{
			Name: p.namer.GoNameToProtoName(t.Name),
			Kind: typesKindProtobuf,

			CommentLines: t.CommentLines,
		}
		p.tracker.AddType(t)
		return t, nil
	}
	return nil, errUnrecognizedType
}

type bodyGen struct {
	locator      ProtobufLocator
	localPackage types.Name

	t *types.Type
}

func (b bodyGen) unknown(sw *generator.SnippetWriter) {
	sw.Do("// Not sure how to generate $.Name$\n", b.t)
}

func (b bodyGen) doStruct(sw *generator.SnippetWriter) {
	if len(b.t.Name.Name) == 0 {
		return
	}
	if isPrivateGoName(b.t.Name.Name) {
		return
	}

	var fields []protoField
	options := []string{}
	allOptions := types.ExtractCommentTags("+", b.t.CommentLines)
	for k, v := range allOptions {
		switch {
		case strings.HasPrefix(k, "genprotoidl.options."):
			key := strings.TrimPrefix(k, "genprotoidl.options.")
			switch key {
			case "marshal":
				if v == "false" {
					options = append(options,
						"(gogoproto.marshaler) = false",
						"(gogoproto.unmarshaler) = false",
						"(gogoproto.sizer) = false",
					)
				}
			default:
				options = append(options, fmt.Sprintf("%s = %s", key, v))
			}
		case k == "genprotoidl.embed":
			fields = []protoField{
				{
					Tag:  1,
					Name: v,
					Type: &types.Type{
						Name: types.Name{
							Name:    v,
							Package: b.localPackage.Package,
							Path:    b.localPackage.Path,
						},
					},
				},
			}
		}
	}

	if fields == nil {
		memberFields, err := membersToFields(b.locator, b.t, b.localPackage)
		if err != nil {
			sw.Do(fmt.Sprintf("// ERROR: type $.Name$ cannot be converted to protobuf: %v\n", err), b.t)
			return
		}
		fields = memberFields
	}

	out := sw.Out()
	genComment(out, b.t.CommentLines, "")
	sw.Do(`message $.Name.Name$ {
`, b.t)

	if len(options) > 0 {
		sort.Sort(sort.StringSlice(options))
		for _, s := range options {
			fmt.Fprintf(out, "  option %s;\n", s)
		}
		fmt.Fprintln(out)
	}

	for i, field := range fields {
		genComment(out, field.CommentLines, "  ")
		fmt.Fprintf(out, "  ")
		switch {
		case field.Map:
		case field.Repeated:
			fmt.Fprintf(out, "repeated ")
		case field.Optional:
			fmt.Fprintf(out, "optional ")
		default:
			fmt.Fprintf(out, "required ")
		}
		sw.Do(`$.Type|local$ $.Name$ = $.Tag$`, field)
		if len(field.Extras) > 0 {
			fmt.Fprintf(out, " [")
			first := true
			for k, v := range field.Extras {
				if first {
					first = false
				} else {
					fmt.Fprintf(out, ", ")
				}
				fmt.Fprintf(out, "%s = %s", k, v)
			}
			fmt.Fprintf(out, "]")
		}
		fmt.Fprintf(out, ";\n")
		if i != len(fields)-1 {
			fmt.Fprintf(out, "\n")
		}
	}
	fmt.Fprintf(out, "}\n\n")
}

type protoField struct {
	LocalPackage types.Name

	Tag      int
	Name     string
	Type     *types.Type
	Map      bool
	Repeated bool
	Optional bool
	Nullable bool
	Extras   map[string]string

	CommentLines string

	OptionalSet bool
}

var (
	errUnrecognizedType = fmt.Errorf("did not recognize the provided type")
)

func isFundamentalProtoType(t *types.Type) (*types.Type, bool) {
	// switch {
	// case t.Kind == types.Struct && t.Name == types.Name{Package: "time", Name: "Time"}:
	// 	return &types.Type{
	// 		Kind: typesKindProtobuf,
	// 		Name: types.Name{Path: "google/protobuf/timestamp.proto", Package: "google.protobuf", Name: "Timestamp"},
	// 	}, true
	// }
	switch t.Kind {
	case types.Slice:
		if t.Elem.Name.Name == "byte" && len(t.Elem.Name.Package) == 0 {
			return &types.Type{Name: types.Name{Name: "bytes"}, Kind: typesKindProtobuf}, true
		}
	case types.Builtin:
		switch t.Name.Name {
		case "string", "uint32", "int32", "uint64", "int64", "bool":
			return &types.Type{Name: types.Name{Name: t.Name.Name}, Kind: typesKindProtobuf}, true
		case "int":
			return &types.Type{Name: types.Name{Name: "int64"}, Kind: typesKindProtobuf}, true
		case "uint":
			return &types.Type{Name: types.Name{Name: "uint64"}, Kind: typesKindProtobuf}, true
		case "float64", "float":
			return &types.Type{Name: types.Name{Name: "double"}, Kind: typesKindProtobuf}, true
		case "float32":
			return &types.Type{Name: types.Name{Name: "float"}, Kind: typesKindProtobuf}, true
		case "uintptr":
			return &types.Type{Name: types.Name{Name: "uint64"}, Kind: typesKindProtobuf}, true
		}
		// TODO: complex?
	}
	return t, false
}

func memberTypeToProtobufField(locator ProtobufLocator, field *protoField, t *types.Type) error {
	// TODO: should locator.TypeFor(t) return an error?
	// if newT, ok := isProtoType(t); ok {
	// 	t = newT
	// }
	var err error
	switch t.Kind {
	case typesKindProtobuf:
		field.Type, err = locator.ProtoTypeFor(t)
	case types.Builtin:
		field.Type, err = locator.ProtoTypeFor(t)
	case types.Map:
		valueField := &protoField{}
		if err := memberTypeToProtobufField(locator, valueField, t.Elem); err != nil {
			return err
		}
		keyField := &protoField{}
		if err := memberTypeToProtobufField(locator, keyField, t.Key); err != nil {
			return err
		}
		field.Type = &types.Type{
			Elem: valueField.Type,
			Key:  keyField.Type,
			Kind: types.Map,
		}
		if !strings.HasPrefix(t.Name.Name, "map[") {
			field.Extras["(gogoproto.casttype)"] = strconv.Quote(locator.CastTypeName(t.Name))
		}
		if _, ok := keyField.Extras["(gogoproto.casttype)"]; ok {
			log.Printf("%s had key with cast type: %#v", t.Name, field)
		}
		field.Map = true
	case types.Pointer:
		if err := memberTypeToProtobufField(locator, field, t.Elem); err != nil {
			return err
		}
		field.Nullable = true
	case types.Alias:
		if err := memberTypeToProtobufField(locator, field, t.Underlying); err != nil {
			log.Printf("failed to alias: %s %s: err", t.Name, t.Underlying.Name, err)
			return err
		}
		if field.Extras == nil {
			field.Extras = make(map[string]string)
		}
		field.Extras["(gogoproto.casttype)"] = strconv.Quote(locator.CastTypeName(t.Name))
		if t.Underlying.Kind == types.Map {
			log.Printf("alias to map: %s %s", t.Name, t.Underlying.Name)
		}
	case types.Slice:
		if t.Elem.Name.Name == "byte" && len(t.Elem.Name.Package) == 0 {
			field.Type = &types.Type{Name: types.Name{Name: "bytes"}, Kind: typesKindProtobuf}
			return nil
		}
		if err := memberTypeToProtobufField(locator, field, t.Elem); err != nil {
			return err
		}
		field.Repeated = true
	case types.Struct:
		if len(t.Name.Name) == 0 {
			return errUnrecognizedType
		}
		field.Type, err = locator.ProtoTypeFor(t)
		field.Nullable = false
	default:
		return errUnrecognizedType
	}
	return err
}

// protobufTagToField extracts information from an existing protobuf tag
// TODO: take a current package
func protobufTagToField(tag string, field *protoField, m types.Member, t *types.Type, localPackage types.Name) error {
	if len(tag) == 0 {
		return nil
	}

	// protobuf:"bytes,3,opt,name=Id,customtype=github.com/gogo/protobuf/test.Uuid"
	parts := strings.Split(tag, ",")
	if len(parts) < 3 {
		return fmt.Errorf("member %q of %q malformed 'protobuf' tag, not enough segments\n", m.Name, t.Name)
	}
	protoTag, err := strconv.Atoi(parts[1])
	if err != nil {
		return fmt.Errorf("member %q of %q malformed 'protobuf' tag, field ID is %q which is not an integer: %v\n", m.Name, t.Name, parts[1], err)
	}
	field.Tag = protoTag
	// TODO: we are converting a Protobuf type back into an internal type, which is questionable
	if last := strings.LastIndex(parts[0], "."); last != -1 {
		prefix := parts[0][:last]
		field.Type = &types.Type{
			Name: types.Name{
				Name:    parts[0][last+1:],
				Package: prefix,
				// TODO: this probably needs to be a lookup into a namer
				Path: strings.Replace(prefix, ".", "/", -1),
			},
			Kind: typesKindProtobuf,
		}
	} else {
		field.Type = &types.Type{
			Name: types.Name{
				Name:    parts[0],
				Package: localPackage.Package,
				Path:    localPackage.Path,
			},
			Kind: typesKindProtobuf,
		}
	}
	switch parts[2] {
	case "rep":
		field.Repeated = true
	case "opt":
		field.Optional = true
	case "req":
	default:
		return fmt.Errorf("member %q of %q malformed 'protobuf' tag, field mode is %q not recognized\n", m.Name, t.Name, parts[2])
	}
	field.OptionalSet = true

	protoExtra := make(map[string]string)
	for i, extra := range parts[3:] {
		parts := strings.SplitN(extra, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("member %q of %q malformed 'protobuf' tag, tag %d should be key=value, got %q\n", m.Name, t.Name, i+4, extra)
		}
		protoExtra[parts[0]] = parts[1]
	}

	field.Extras = protoExtra
	if name, ok := protoExtra["name"]; ok {
		field.Name = name
		delete(protoExtra, "name")
	}

	return nil
}

func membersToFields(locator ProtobufLocator, t *types.Type, localPackage types.Name) ([]protoField, error) {
	fields := []protoField{}

	for _, m := range t.Members {
		if isPrivateGoName(m.Name) {
			// skip private fields
			continue
		}
		tags := reflect.StructTag(m.Tags)
		field := protoField{
			LocalPackage: localPackage,

			Tag:    -1,
			Extras: make(map[string]string),
		}

		if err := protobufTagToField(tags.Get("protobuf"), &field, m, t, localPackage); err != nil {
			return nil, err
		}

		// extract information from JSON field tag
		if tag := tags.Get("json"); len(tag) > 0 {
			parts := strings.Split(tag, ",")
			if len(field.Name) == 0 && len(parts[0]) != 0 {
				field.Name = parts[0]
			}
			/*if len(parts) > 1 {
				for _, s := range parts[1:] {
					switch s {
					case "omitempty":
						// TODO: make this nullable
						//if !field.OptionalSet {
						//	field.Optional = true
						//	field.OptionalSet = true
						//}
					case "inline":
						// TODO: inline all members, give them contextual tag that is non-conflicting
					}
				}
			}*/
		}

		if field.Type == nil {
			if err := memberTypeToProtobufField(locator, &field, m.Type); err != nil {
				return nil, fmt.Errorf("unable to embed type %q as field %q in %q: %v", m.Type, field.Name, t.Name, err)
			}
		}
		if len(field.Name) == 0 {
			field.Name = strings.ToLower(m.Name[:1]) + m.Name[1:]
		}

		if field.Map && field.Repeated {
			// maps cannot be repeated
			field.Repeated = false
			field.Nullable = true
		}
		// embedded fields that are not repeated should be considered required
		//if m.Embedded && !field.Repeated {
		//	field.Nullable = false
		//}

		if !field.Nullable {
			field.Extras["(gogoproto.nullable)"] = "false"
		}
		if (field.Type.Name.Name == "bytes" && field.Type.Name.Package == "") || (field.Repeated && field.Type.Name.Package == "" && isPrivateGoName(field.Type.Name.Name)) {
			delete(field.Extras, "(gogoproto.nullable)")
		}
		if field.Name != m.Name {
			field.Extras["(gogoproto.customname)"] = strconv.Quote(m.Name)
		}
		field.CommentLines = m.CommentLines
		fields = append(fields, field)
	}

	// assign tags
	highest := 0
	byTag := make(map[int]*protoField)
	// fields are in Go struct order, which we preserve
	for i := range fields {
		field := &fields[i]
		tag := field.Tag
		if tag != -1 {
			if existing, ok := byTag[tag]; ok {
				return nil, fmt.Errorf("field %q and %q in %q both have tag %d", field.Name, existing.Name, tag)
			}
			byTag[tag] = field
		}
		if tag > highest {
			highest = tag
		}
	}
	// starting from the highest observed tag, assign new field tags
	for i := range fields {
		field := &fields[i]
		if field.Tag != -1 {
			continue
		}
		highest++
		field.Tag = highest
		byTag[field.Tag] = field
	}
	return fields, nil
}

func genComment(out io.Writer, comment, indent string) {
	lines := strings.Split(comment, "\n")
	for {
		l := len(lines)
		if l == 0 || len(lines[l-1]) != 0 {
			break
		}
		lines = lines[:l-1]
	}
	for _, c := range lines {
		fmt.Fprintf(out, "%s// %s\n", indent, c)
	}
}

type protoIDLFileType struct{}

func (ft protoIDLFileType) AssembleFile(f *generator.File, pathname string) error {
	log.Printf("Assembling IDL file %q", pathname)
	destFile, err := os.Create(pathname)
	if err != nil {
		return err
	}
	defer destFile.Close()

	b := &bytes.Buffer{}
	et := generator.NewErrorTracker(b)
	ft.assemble(et, f)
	if et.Error() != nil {
		return et.Error()
	}

	// TODO: is there an IDL formatter?
	_, err = destFile.Write(b.Bytes())
	return err
}

func (ft protoIDLFileType) assemble(w io.Writer, f *generator.File) {
	w.Write(f.Header)

	fmt.Fprint(w, "syntax = 'proto2';\n\n")

	if len(f.PackageName) > 0 {
		fmt.Fprintf(w, "package %v;\n\n", f.PackageName)
	}

	if len(f.Imports) > 0 {
		for i := range f.Imports {
			fmt.Fprintf(w, "import %q;\n", i)
		}
		fmt.Fprint(w, "\n")
	}

	if f.Vars.Len() > 0 {
		fmt.Fprintf(w, "%s\n", f.Vars.String())
	}

	w.Write(f.Body.Bytes())
}

func isPackable(t *types.Type) bool {
	if t.Kind != typesKindProtobuf {
		return false
	}
	switch t.Name.Name {
	case "int32", "int64", "varint":
		return true
	default:
		return false
	}
}

func isPrivateGoName(name string) bool {
	if len(name) == 0 {
		return true
	}
	return strings.ToLower(name[:1]) == name[:1]
}
