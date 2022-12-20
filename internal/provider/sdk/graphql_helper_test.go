//go:build acceptance
// +build acceptance

package sdk

import (
	"context"
	"log"
	"testing"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/internal/provider/testutil"
)

func TestAcc_GraphQL_basic(t *testing.T) {

	query := GraphQLQuery{
		`query {currentUser {name, bot, gitpodEnabled, groupCount, id, namespace{id}, publicEmail, username}}`,
	}

	var response CurrentUserResponse
	_, err := SendGraphQLRequest(context.Background(), testutil.TestGitlabClient, query, &response)
	if err != nil {
		log.Println(err)
		t.Fail()
	}

	if response.Data.CurrentUser.Name != "Administrator" {
		t.Fail()
	}
}
