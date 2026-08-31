package clients

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcConn wraps the gRPC connections to all backing services.
type GrpcConn struct {
	Auth          *grpc.ClientConn
	User          *grpc.ClientConn
	Question      *grpc.ClientConn
	Matching      *grpc.ClientConn
	Solution      *grpc.ClientConn
	Chat          *grpc.ClientConn
	Notification  *grpc.ClientConn
	Knowledgebase *grpc.ClientConn
	Analytics     *grpc.ClientConn
	Admin         *grpc.ClientConn
}

func dial(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(10*1024*1024)),
	)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}
	return conn, nil
}

// New dials all services and returns the connection set.
func New(addrs Addrs) (*GrpcConn, error) {
	g := &GrpcConn{}
	var err error

	if g.Auth, err = dial(addrs.Auth); err != nil {
		return nil, err
	}
	if g.User, err = dial(addrs.User); err != nil {
		return nil, err
	}
	if g.Question, err = dial(addrs.Question); err != nil {
		return nil, err
	}
	if g.Matching, err = dial(addrs.Matching); err != nil {
		return nil, err
	}
	if g.Solution, err = dial(addrs.Solution); err != nil {
		return nil, err
	}
	if g.Chat, err = dial(addrs.Chat); err != nil {
		return nil, err
	}
	if g.Notification, err = dial(addrs.Notification); err != nil {
		return nil, err
	}
	if g.Knowledgebase, err = dial(addrs.Knowledgebase); err != nil {
		return nil, err
	}
	if g.Analytics, err = dial(addrs.Analytics); err != nil {
		return nil, err
	}
	if g.Admin, err = dial(addrs.Admin); err != nil {
		return nil, err
	}

	return g, nil
}

func (g *GrpcConn) Close() {
	closeConn(g.Auth)
	closeConn(g.User)
	closeConn(g.Question)
	closeConn(g.Matching)
	closeConn(g.Solution)
	closeConn(g.Chat)
	closeConn(g.Notification)
	closeConn(g.Knowledgebase)
	closeConn(g.Analytics)
	closeConn(g.Admin)
}

func closeConn(c *grpc.ClientConn) {
	if c != nil {
		_ = c.Close()
	}
}
