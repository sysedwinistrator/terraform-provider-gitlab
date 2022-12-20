package main

import (
	"os"

	"golang.org/x/tools/go/analysis/singlechecker"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/tools/passes/apiunused"
)

func main() {
	apiunused.Output = os.Stdout
	singlechecker.Main(apiunused.Analyzer)
}
