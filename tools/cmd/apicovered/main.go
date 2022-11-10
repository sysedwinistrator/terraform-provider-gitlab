package main

import (
	"os"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/tools/passes/apicovered"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	apicovered.Output = os.Stdout
	singlechecker.Main(apicovered.Analyzer)
}
