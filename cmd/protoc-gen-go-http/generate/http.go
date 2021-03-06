package generate

import (
	"fmt"
	"github.com/waldenlake/go-kit/cmd/protoc-gen-go-http/template"
	"google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"strings"
)

const (
	version              = "v0.0.1"
	contextPackage       = protogen.GoImportPath("context")
	ginPackage           = protogen.GoImportPath("github.com/gin-gonic/gin")
	transportHTTPPackage = protogen.GoImportPath("github.com/waldenlake/go-kit/transport/http")
	deprecationComment   = "// Deprecated: Do not use."
)

var methodSets = make(map[string]int)

func GenerateFIle(plugin *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	if len(file.Services) == 0 || (!hasHTTPRule(file.Services)) {
		return nil
	}
	filename := file.GeneratedFilenamePrefix + "_router.pb.go"
	gfile := plugin.NewGeneratedFile(filename, file.GoImportPath)
	gfile.P("// Code generated by protoc-gen-go-router. DO NOT EDIT.")
	gfile.P("// versions:")
	gfile.P(fmt.Sprintf("// protoc-gen-go-router %s", version))
	gfile.P()
	gfile.P("package ", file.GoPackageName)
	gfile.P()
	generateFileContent(plugin, file, gfile)
	return gfile
}

func generateFileContent(plugin *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile) {
	if len(file.Services) == 0 {
		return
	}
	g.P("// This is a compile-time assertion to ensure that this generated file")
	g.P("// is compatible with the go-kit package it is being compiled against.")
	g.P("var _ = new(", contextPackage.Ident("Context"), ")")
	g.P("var _ = ", ginPackage.Ident("Version"))
	g.P("var _ = ", transportHTTPPackage.Ident("JSON"))
	g.P()

	for _, service := range file.Services {
		genService(plugin, file, g, service)
	}
}

func genService(plugin *protogen.Plugin, file *protogen.File, g *protogen.GeneratedFile, service *protogen.Service) {
	if service.Desc.Options().(*descriptorpb.ServiceOptions).GetDeprecated() {
		g.P("//")
		g.P(deprecationComment)
	}
	sd := &template.ServiceDesc{
		ServiceType: service.GoName,
		ServiceName: string(service.Desc.FullName()),
		Metadata:    file.Desc.Path(),
	}
	for _, method := range service.Methods {
		if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
			continue
		}
		rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
		if rule != nil && ok {
			for _, bind := range rule.AdditionalBindings {
				sd.Methods = append(sd.Methods, buildHTTPRule(g, method, bind))
			}
			sd.Methods = append(sd.Methods, buildHTTPRule(g, method, rule))
		}
	}
	if len(sd.Methods) != 0 {
		g.P(sd.Execute())
	}
}

func buildHTTPRule(g *protogen.GeneratedFile, m *protogen.Method, rule *annotations.HttpRule) *template.MethodDesc {
	var (
		path         string
		method       string
		body         string
		responseBody string
	)
	switch pattern := rule.Pattern.(type) {
	case *annotations.HttpRule_Get:
		path = pattern.Get
		method = "GET"
	case *annotations.HttpRule_Put:
		path = pattern.Put
		method = "PUT"
	case *annotations.HttpRule_Post:
		path = pattern.Post
		method = "POST"
	case *annotations.HttpRule_Delete:
		path = pattern.Delete
		method = "DELETE"
	case *annotations.HttpRule_Patch:
		path = pattern.Patch
		method = "PATCH"
	case *annotations.HttpRule_Custom:
		path = pattern.Custom.Path
		method = pattern.Custom.Kind
	}
	body = rule.Body
	responseBody = rule.ResponseBody
	md := buildMethodDesc(g, m, method, path)
	if method == "GET" {
		md.HasBody = false
	} else if body == "*" {
		md.HasBody = true
		md.Body = ""
	} else if body != "" {
		md.HasBody = true
		md.Body = "." + camelCaseVars(body)
	} else {
		md.HasBody = false
	}
	if responseBody == "*" {
		md.ResponseBody = ""
	} else if responseBody != "" {
		md.ResponseBody = "." + camelCaseVars(responseBody)
	}
	return md
}

func buildMethodDesc(g *protogen.GeneratedFile, m *protogen.Method, method, path string) *template.MethodDesc {
	defer func() { methodSets[m.GoName]++ }()
	return &template.MethodDesc{
		Name:    m.GoName,
		Num:     methodSets[m.GoName],
		Request: g.QualifiedGoIdent(m.Input.GoIdent),
		Reply:   g.QualifiedGoIdent(m.Output.GoIdent),
		Path:    path,
		Method:  method,
		HasVars: len(buildPathVars(m, path)) > 0,
	}
}

func hasHTTPRule(services []*protogen.Service) bool {
	for _, service := range services {
		for _, method := range service.Methods {
			if method.Desc.IsStreamingClient() || method.Desc.IsStreamingServer() {
				continue
			}
			rule, ok := proto.GetExtension(method.Desc.Options(), annotations.E_Http).(*annotations.HttpRule)
			if rule != nil && ok {
				return true
			}
		}
	}
	return false
}

func buildPathVars(method *protogen.Method, path string) (res []string) {
	for _, v := range strings.Split(path, "/") {
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			name := strings.TrimRight(strings.TrimLeft(v, "{"), "}")
			res = append(res, name)
		}
	}
	return
}

func camelCaseVars(s string) string {
	/*var (
		vars []string
		subs = strings.Split(s, ".")
	)
	for _, sub := range subs {
		vars = append(vars, camelCase(sub))
	}
	return strings.Join(vars, ".")*/
	return ""
}
