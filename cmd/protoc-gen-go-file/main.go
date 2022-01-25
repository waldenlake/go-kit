package main

import (
	"flag"
	"github.com/walden/go-kit/protoc-gen-go-file/generate"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

func main() {
	protogen.Options{
		ParamFunc: flag.CommandLine.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		for _, file := range plugin.Files {
			if file.Generate {
				generate.GenerateFIle(plugin, file)
			}
		}
		return nil
	})
}
