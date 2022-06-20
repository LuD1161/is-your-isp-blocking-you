package cmd

const (
	// Error Code constants
	CONN_RESET   = 401
	CONN_BLOCKED = 402
	CONN_UNKNOWN = 500
	CONN_TIMEOUT = 501
	NO_SUCH_HOST = 404
	OTHER_ERROR  = -1

	// Censorship Methods
	NOT_FILTERED   = 1000
	HTTP_FILTERING = 1001
	DNS_FILTERING  = 1002
	IP_FILTERING   = 1003
	SNI_FILTERING  = 1004

	// HTTP Client Constants
	MAX_RETRIES = 3
)