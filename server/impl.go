package main

import (
	"context"
	"fmt"
	"server/gen"
	helloworldpb "server/gen"

	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	helloworldpb.UnimplementedGreeterServer
	dbpool *pgxpool.Pool
}

func NewServer(dbpool *pgxpool.Pool) *server {
	return &server{
		dbpool: dbpool,
	}
}

func (s *server) SayHello(ctx context.Context, in *helloworldpb.HelloRequest) (*helloworldpb.HelloReply, error) {
	user, ok := GetUserFromContext(ctx)
	if !ok {
		fmt.Println("NOT OK")
		return nil, errUnauthenticated
	}

	tx, err := s.dbpool.Begin(ctx)
	if err != nil {
		return nil, errUnauthenticated
	}
	cases, err := user.GetHibpCases(ctx, tx)
	if err != nil {
		return nil, errUnauthenticated
	}

	out := make([]*gen.HibpCase, 0, len(cases))
	for _, c := range cases {
		a := gen.HibpCase{
			Name: c.Name,
		}
		out = append(out, &a)
	}

	return &helloworldpb.HelloReply{HibpCases: out}, nil
}
