package main

import (
	"os"

	"golang.org/x/tools/go/analysis/singlechecker"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/tools/passes/apicovered"
)

func main() {
	apicovered.Output = os.Stdout
	singlechecker.Main(apicovered.Analyzer)
}
