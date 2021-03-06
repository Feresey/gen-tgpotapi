{{.Head}}

const Version = "{{.Version}}"

// TODO: category description
{{range $typename, $desc := .Types}} {{- if not (skip $typename)}}

// {{camel $typename}}
// {{format $desc.Description.PlainText 0}}
type {{camel $typename}} struct {
{{- range $idx, $put_required := $.RequiredOrder}}
	{{- range $field_name, $field_desc := $desc.Fields}}
		{{- $type := get_type $field_name $typename $field_desc.Types }}

		{{- if eq $field_desc.Required $put_required}}
		// {{camel $field_name}}
		// {{format $field_desc.Description.PlainText 1}}
		{{camel $field_name}} {{if and (not $type.IsArray) (not $field_desc.Required) -}}
				*{{end -}}
			{{with $type.GoType}}{{if eq . "InputFile"}}FileID{{else}}{{.}}{{end}}{{end -}}
			`json:"{{$field_name}}{{if not $field_desc.Required}},omitempty{{end}}"`
		{{- end}}
	{{- end}}
{{- end}}
}

{{range $field_name, $field_desc := $desc.Fields}}
{{- $type := get_type $field_name $typename $field_desc.Types }}
{{- if not $type.IsArray}}
{{- $required := $field_desc.Required}}
{{- $simple := $type.IsSimpleType}}
{{- $ttype := $type.GoType}}
{{- if eq $ttype "InputFile"}}
{{- $ttype = "FileID"}}
{{- end}}

func (t *{{camel $typename}}) Get{{camel $field_name}}() {{if not $simple}}*{{end}}{{$ttype}} {
	{{- if $simple}}
	var res {{$ttype}}
	{{- end}}
	if t == nil {
		return {{if $simple}}res{{else}}nil{{end}}
	}
	{{- if or $required $type.IsArray}}
	return {{if and $required (not $simple)}}&{{end}}t.{{camel $field_name}}
	{{- else}}
	{{- if and (not $required) (not $simple) }}
	return t.{{camel $field_name}}
	{{- else}}
	if field := t.{{camel $field_name}}; field != nil {
		return {{if $simple}}*{{end}}field
	}
	return {{if $simple}}res{{else}}nil{{end}}
	{{- end}}
	{{- end}}
}
{{- end}}
{{- end}}

{{end}}{{end}}