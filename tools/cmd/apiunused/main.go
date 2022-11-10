package main

import (
	"os"

	"gitlab.com/gitlab-org/terraform-provider-gitlab/tools/passes/apiunused"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	apiunused.Output = os.Stdout
	singlechecker.Main(apiunused.Analyzer)
}
