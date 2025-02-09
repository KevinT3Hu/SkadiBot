package utils

import (
	"google.golang.org/grpc/resolver"
)

type Doc2VecGrpcResolver struct {
	destination string
	cc          resolver.ClientConn
}

func NewDoc2VecGrpcResolver(cc resolver.ClientConn) *Doc2VecGrpcResolver {
	r := &Doc2VecGrpcResolver{
		destination: GetConfig().Doc2VecConfig.Destination,
		cc:          cc,
	}
	go r.start(Doc2VecAddrChan)
	return r
}

func (r *Doc2VecGrpcResolver) start(addrChan <-chan string) {
	for addr := range addrChan {
		r.cc.UpdateState(resolver.State{Addresses: []resolver.Address{{Addr: addr}}})
	}
}

func (r *Doc2VecGrpcResolver) ResolveNow(resolver.ResolveNowOptions) {
}

func (r *Doc2VecGrpcResolver) Close() {
}

type Doc2VecGrpcResolverBuilder struct{}

func (d *Doc2VecGrpcResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	return NewDoc2VecGrpcResolver(cc), nil
}

func (d *Doc2VecGrpcResolverBuilder) Scheme() string {
	return "doc2vec"
}
