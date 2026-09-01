package clients

// Addrs holds the gRPC addresses of all backing services.
type Addrs struct {
	Auth          string
	User          string
	Doubt         string
	Matching      string
	Knowledgebase string
	Analytics     string
	Admin         string
}
