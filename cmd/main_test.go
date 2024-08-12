package main

import "testing"

func TestGenerateProtoFile(t *testing.T) {
	s := "./protobuf"
	dst := "gen/go"
	files := []string{"account_service.proto", "enums.proto"}
	generateProtoFile(s, dst, files...)
}
