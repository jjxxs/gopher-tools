package error

import "errors"

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
)

func Of(errMsg string) error {
	return errors.New(errMsg)
}
