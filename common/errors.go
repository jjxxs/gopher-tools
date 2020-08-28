package common

import "errors"

const (
	JsonFailedToMarshal   = "failed to marshal json"
	JsonFailedToUnmarshal = "failed to unmarshal json"

	DatabaseFailedToConnect = "failed to connect to database"
	DatabaseFailedToQuery   = "failed to query database"
	DatabaseFailedToExecute = "failed to execute on database"
	DatabaseFailedToScanRow = "failed to scan row of database"
)

func Of(errMsg string) error {
	return errors.New(errMsg)
}
