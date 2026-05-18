package client

// InternalServiceClient provides clients for calling other internal services
type InternalServiceClient struct {
	// TODO: Add HTTP clients for each internal service
	// userClient    *UserServiceClient
	// briefClient *BriefServiceClient
	// videoClient *VideoServiceClient
}

// NewInternalServiceClient creates a new internal service client
func NewInternalServiceClient() *InternalServiceClient {
	return &InternalServiceClient{}
}