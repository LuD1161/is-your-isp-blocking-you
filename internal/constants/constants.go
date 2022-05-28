package constants

const (
	// Error Code constants
	CONN_OK      = 1
	CONN_RESET   = 401
	CONN_BLOCKED = 402
	CONN_UNKNOWN = 500
	CONN_TIMEOUT = 501
	NO_SUCH_HOST = 404

	// HTTP Client Constants
	MAX_RETRIES         = 3
	HTTP_CLIENT_TIMEOUT = 5

	// Program constants
	MAX_GOROUTINES = 1000
)
