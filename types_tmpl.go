// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package gowsdl

var typesTmpl = `
{{define "SimpleType"}}
	{{$typeName := toGoType .Name false | removePointerFromType}}
	{{if .Doc}} {{.Doc | comment}} {{end}}
	{{if ne .List.ItemType ""}}
		type {{$typeName}} []{{toGoType .List.ItemType false | removePointerFromType}}
	{{else if ne .Union.MemberTypes ""}}
		type {{$typeName}} string
	{{else if .Union.SimpleType}}
		type {{$typeName}} string
	{{else if .Restriction.Base}}
		type {{$typeName}} {{toGoType .Restriction.Base false | removePointerFromType}}
    {{else}}
		type {{$typeName}} interface{}
	{{end}}

	{{if .Restriction.Enumeration}}
	const (
		{{with .Restriction}}
			{{range .Enumeration}}
				{{if .Doc}} {{.Doc | comment}} {{end}}
				{{$typeName}}{{$value := replaceReservedWords .Value}}{{$value | makePublic}} {{$typeName}} = "{{goString .Value}}" {{end}}
		{{end}}
	)
	{{end}}
{{end}}

{{define "ComplexContent"}}
	{{$baseType := toGoType .Extension.Base false}}
	{{ if $baseType }}
		{{$baseType}}
	{{end}}

	{{template "Elements" .Extension.Sequence}}
	{{template "Elements" .Extension.Choice}}
	{{template "Elements" .Extension.SequenceChoice}}
	{{template "Attributes" .Extension.Attributes}}
{{end}}

{{define "Attributes"}}
    {{ $targetNamespace := getNS }}
	{{range .}}
		{{if .Doc}} {{.Doc | comment}} {{end}}
		{{ if ne .Type "" }}
			{{ normalize .Name | makeFieldPublic}} {{toGoType .Type false}} ` + "`" + `xml:"{{with $targetNamespace}}{{.}} {{end}}{{.Name}},attr,omitempty" json:"{{.Name}},omitempty"` + "`" + `
		{{ else }}
			{{ normalize .Name | makeFieldPublic}} string ` + "`" + `xml:"{{with $targetNamespace}}{{.}} {{end}}{{.Name}},attr,omitempty" json:"{{.Name}},omitempty"` + "`" + `
		{{ end }}
	{{end}}
{{end}}

{{define "SimpleContent"}}
	Value {{toGoType .Extension.Base false}} ` + "`xml:\",chardata\" json:\"-,\"`" + `
	{{template "Attributes" .Extension.Attributes}}
{{end}}

{{define "ComplexTypeInline"}}
	{{replaceReservedWords .Name | makePublic}} {{if eq .MaxOccurs "unbounded"}}[]{{end}}struct {
	{{with .ComplexType}}
		{{if ne .ComplexContent.Extension.Base ""}}
			{{template "ComplexContent" .ComplexContent}}
		{{else if ne .SimpleContent.Extension.Base ""}}
			{{template "SimpleContent" .SimpleContent}}
		{{else}}
			{{template "Elements" .Sequence}}
			{{template "Elements" .Choice}}
			{{template "Elements" .SequenceChoice}}
			{{template "Elements" .All}}
			{{template "Attributes" .Attributes}}
		{{end}}
	{{end}}
	} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
{{end}}

{{define "Elements"}}
	{{range .}}
		{{if ne .Ref ""}}
			{{removeNS .Ref | replaceReservedWords  | makePublic}} {{if eq .MaxOccurs "unbounded"}}[]{{end}}{{toGoType .Ref .Nillable }} ` + "`" + `xml:"{{.Ref | removeNS}},omitempty" json:"{{.Ref | removeNS}},omitempty"` + "`" + `
		{{else}}
		{{if not .Type}}
			{{if .SimpleType}}
				{{if .Doc}} {{.Doc | comment}} {{end}}
				{{if ne .SimpleType.List.ItemType ""}}
					{{ normalize .Name | makeFieldPublic}} []{{toGoType .SimpleType.List.ItemType false}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{else}}
					{{ normalize .Name | makeFieldPublic}} {{toGoType .SimpleType.Restriction.Base false}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + `
				{{end}}
			{{else}}
				{{template "ComplexTypeInline" .}}
			{{end}}
		{{else}}
			{{$goTypeName := toGoType .Type (eq .MinOccurs "0")}}

			{{if isAbstract .Type}}
				{{$goTypeName = print $goTypeName "Wrapper"}}
			{{end}}

			{{if .Doc}}{{.Doc | comment}} {{end}}
			{{replaceAttrReservedWords .Name | makeFieldPublic}} {{if eq .MaxOccurs "unbounded"}}[]{{end}}{{$goTypeName}} ` + "`" + `xml:"{{.Name}},omitempty" json:"{{.Name}},omitempty"` + "`" + ` {{end}}
		{{end}}
	{{end}}
{{end}}

{{define "Any"}}
	{{range .}}
		Items     []string ` + "`" + `xml:",any" json:"items,omitempty"` + "`" + `
	{{end}}
{{end}}

{{range .Schemas}}
	{{ $targetNamespace := setNS .TargetNamespace }}

	{{range .SimpleType}}
		{{template "SimpleType" .}}
	{{end}}

	{{range .ComplexTypes}}
		{{/* ComplexTypeGlobal */}}
		{{$typeName := toGoType .Name false | removePointerFromType}}
		{{if and (eq (len .SimpleContent.Extension.Attributes) 0) (eq (toGoType .SimpleContent.Extension.Base false) "string") }}
			type {{$typeName}} string
		{{else}}
			type {{$typeName}} struct {
				{{$type := findNameByType .Name}}
				{{if ne .Name $type}}
					XMLName xml.Name ` + "`xml:\"{{$targetNamespace}} {{$type}}\"`" + `
				{{end}}

				{{if ne .ComplexContent.Extension.Base ""}}
					{{template "ComplexContent" .ComplexContent}}
				{{else if ne .SimpleContent.Extension.Base ""}}
					{{template "SimpleContent" .SimpleContent}}
				{{else}}
					{{template "Elements" .Sequence}}
					{{template "Any" .Any}}
					{{template "Elements" .Choice}}
					{{template "Elements" .SequenceChoice}}
					{{template "Elements" .All}}
					{{template "Attributes" .Attributes}}
				{{end}}
			}
		{{end}}

		{{/* Abstract types will be wrapped in a wrapper type for XML marshaling */}}
		{{/* This is however only done if the abstract type is used anywhere. If only the extended types are used there is no need for abstract type to be generated. */}}
		{{if and (eq .Abstract true) (ne (findNameByType .Name) .Name)}}
			{{$exts := getExtentions .Name}}

			// {{$typeName}}Wrapper is a wrapper for the abstract type {{$typeName}}.
			type {{$typeName}}Wrapper struct {
				{{ range $exts }}
					{{toGoType . false | removePointerFromType}} *{{toGoType . false | replaceReservedWords | makePublic}}
				{{end}}
			}

			// MarshalXML implements the xml.Marshaler interface for {{$typeName}}Wrapper.
			func (w {{$typeName}}Wrapper) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
				start.Name.Local = "{{findNameByType $typeName}}"
				if err := e.EncodeToken(start); err != nil {
					return err
				}

				switch {
				{{range $exts}}
					case w.{{toGoType . false | removePointerFromType}} != nil :
					if err := e.EncodeElement(w.{{toGoType . false | removePointerFromType}}, xml.StartElement{Name: xml.Name{Local: "{{findNameByType .}}"}}); err != nil {
						return err
					}
				{{end}}
				}
				return e.EncodeToken(xml.EndElement{Name: start.Name})
			}

			// UnmarshalXML implements the xml.Unmarshaler interface for {{$typeName}}Wrapper.
			func (w *{{$typeName}}Wrapper) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
				for {
					tok, err := d.Token()
					if err != nil {
						return err
					}
					if el, ok := tok.(xml.StartElement); ok {
						switch el.Name.Local {
						{{range $exts}}
						case "{{findNameByType .}}":
							var item {{toGoType . false | replaceReservedWords | makePublic}}
							if err := d.DecodeElement(&item, &el); err != nil {
								return err
							}
							w.{{toGoType . false | removePointerFromType}} = &item
						{{end}}
						default:
							return fmt.Errorf("unexpected element %s in {{$typeName}}Wrapper", el.Name.Local)
						}
					} else if _, ok := tok.(xml.EndElement); ok {
						break
					}
				}
				return nil
			}
		{{end}}
	{{end}}
{{end}}
`
