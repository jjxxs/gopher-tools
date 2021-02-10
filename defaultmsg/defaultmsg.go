package defaultmsg

const (
	// JSON
	JsonFailedToMarshal   = "failed to marshal json"
	JsonFailedToUnmarshal = "failed to unmarshal json"

	// Database
	DatabaseFailedToConnect = "failed to connect to database"
	DatabaseFailedToQuery   = "failed to query database"
	DatabaseFailedToExecute = "failed to execute on database"
	DatabaseFailedToScanRow = "failed to scan row of database"

	// Requests
	InvalidRequestType  = "invalid request type"
	InvalidResponseType = "invalid response type"

	// Websocket
	WebsocketFailedToUpgradeConnection = "failed to upgrade to websocket-connection"

	// Service
	ServiceStarted = "started service"
	ServiceStopped = "stopped service"
	ServiceFailure = "failed service"

	// Misc
	EnvironmentVariableNotSet  = "environment variable not set"
	EnvironmentVariableInvalid = "environment variable has invalid value"
)
