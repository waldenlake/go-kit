package template

import (
	"bytes"
	"strings"
	"text/template"
)

var fileTemplate = `
{{$svrType := .ServiceType}}
{{$svrName := .ServiceName}}
type {{.ServiceType}} interface {
{{- range .MethodSets}}
	{{.Name}}(c *gin.Context)
{{- end}}
}

func Register{{.ServiceType}}Router(s *http.Server, srv {{.ServiceType}}) {
	r := s.Router()
	{{- range .Methods}}
	r.{{.Method}}("{{.Path}}", {{.Name}}Handler(srv))
	{{- end}}
}

{{range .Methods}}
func {{.Name}}Handler(srv {{$svrType}}) func(c *gin.Context) {
	return func(c *gin.Context) {
		srv.{{.Name}}(c)
	}
}
{{end}}
`

type ServiceDesc struct {
	ServiceType string // Greeter
	ServiceName string // helloworld.Greeter
	Metadata    string // api/helloworld/helloworld.proto
	Methods     []*MethodDesc
	MethodSets  map[string]*MethodDesc
}

type MethodDesc struct {
	Name    string
	Request string
	Reply   string
	Path    string
	Method  string
}

func (s *ServiceDesc) Execute() string {
	s.MethodSets = make(map[string]*MethodDesc)
	for _, m := range s.Methods {
		s.MethodSets[m.Name] = m
	}
	buf := new(bytes.Buffer)
	tmpl, err := template.New("http").Parse(strings.TrimSpace(fileTemplate))
	if err != nil {
		panic(err)
	}
	if err := tmpl.Execute(buf, s); err != nil {
		panic(err)
	}
	return strings.Trim(string(buf.Bytes()), "\r\n")
}
