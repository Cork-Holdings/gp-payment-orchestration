package utils

import "fmt"

const (
	InvReqBody = "Invalid Request Body"
	FailBind   = "Failed to  JSON body"
	UnParse    = "Unable to parse form"
	MissData   = "Data is missing"
	InvalData  = "Invalid Data"
)

func FailedToCreate(resource string) string {
	return fmt.Sprintf("Failed to create %s", resource)
}

func SuccessCreate(resource string) string {
	return fmt.Sprintf("%s created", resource)
}
func FailedToRetrieve(resource string) string {
	return fmt.Sprintf("Failed to retrieve %s", resource)
}

func SuccessfullyRetrieve(resource string) string {
	return fmt.Sprintf("%s retrieved", resource)
}

func FailedToUpdate(resource string) string {
	return fmt.Sprintf("Failed to update %s", resource)
}

func SuccessUpdate(resource string) string {
	return fmt.Sprintf("%s updated", resource)
}

func FailedToDelete(resource string) string {
	return fmt.Sprintf("Failed to delete %s", resource)
}

func SuccessDelete(resource string) string {
	return fmt.Sprintf("%s deleted", resource)
}
