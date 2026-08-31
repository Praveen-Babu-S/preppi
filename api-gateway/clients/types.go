package clients

// Addrs holds the gRPC addresses of all backing services.
type Addrs struct {
	Auth          string
	User          string
	Question      string
	Matching      string
	Solution      string
	Chat          string
	Notification  string
	Knowledgebase string
	Analytics     string
	Admin         string
}
