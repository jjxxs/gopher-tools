package common

import "errors"

const (
	jsonFailedToMarshal   = "failed to marshal json"
	jsonFailedToUnmarshal = "failed to unmarshal json"

	databaseFailedToConnect = "failed to connect to database"
	databaseFailedToQuery   = "failed to query database"
	databaseFailedToExecute = "failed to execute on database"
	databaseFailedToScanRow = "failed to scan row of database"
)

func Of(errMsg string) error {
	return errors.New(errMsg)
}
