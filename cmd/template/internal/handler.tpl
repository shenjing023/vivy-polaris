package internal

import "{{.PkgName}}/{{.GRPCPath}}"

type Server struct {
	pb.Unimplemented{{.ServerName}}Server
}