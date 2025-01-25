package blockchain

type Network struct {
	Segment int
	Routers []*Router
}

type Router struct {
	XchgAddr     string
	CurrentStake uint64
	IpAddr       string
}
