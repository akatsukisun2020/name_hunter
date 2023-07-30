package main

import (
	"context"

	"github.com/akatsukisun2020/name_hunter/service"
	pb "github.com/akatsukisun2020/proto_proj/name_hunter"
)

type NameHunter struct {
	pb.UnimplementedNameHunterHttpServer
}

func (s *NameHunter) RandomName(ctx context.Context, req *pb.RandomNameReq) (*pb.RandomNameRsp, error) {
	return service.RandomName(ctx, req)
}
