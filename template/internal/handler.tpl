package internal

import "{{.PkgName}}/pkg/grpc/pb"

type Server struct {
	pb.Unimplemented{{.ServerName}}Server
}