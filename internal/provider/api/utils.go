package api

import "github.com/xanzy/go-gitlab"

func Is404(err error) bool {
	if errResponse, ok := err.(*gitlab.ErrorResponse); ok &&
		errResponse.Response != nil &&
		errResponse.Response.StatusCode == 404 {
		return true
	}
	return false
}
