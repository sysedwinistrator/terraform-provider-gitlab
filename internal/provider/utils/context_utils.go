package utils

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// The regex to use for masking the token configurations
const gitlabPersonlAccessTokenRegex string = "glpat-.{20}"

// Takes in a provider context, and applies masking to the root and `GitLab` subsystem logger.
// the `GitLab` subsystem is what is used for logging API calls.
func ApplyLogMaskingToContext(ctx context.Context) context.Context {

	// Configure the "root" logger within the context
	// This will mask any logging done directly using the tflog command that doesn't explicitly call a subsystem.
	ctx = tflog.MaskMessageRegexes(ctx, regexp.MustCompile(gitlabPersonlAccessTokenRegex))
	ctx = tflog.MaskLogRegexes(ctx, regexp.MustCompile(gitlabPersonlAccessTokenRegex))
	ctx = tflog.MaskAllFieldValuesRegexes(ctx, regexp.MustCompile(gitlabPersonlAccessTokenRegex))

	// The "GitLab" subsystem is what is used for logging API messages, so configure it here as well.
	subSystemName := "GitLab"
	ctx = tflog.NewSubsystem(ctx, subSystemName)
	ctx = tflog.SubsystemMaskMessageRegexes(ctx, subSystemName, regexp.MustCompile(gitlabPersonlAccessTokenRegex))
	ctx = tflog.SubsystemMaskLogRegexes(ctx, subSystemName, regexp.MustCompile(gitlabPersonlAccessTokenRegex))
	ctx = tflog.SubsystemMaskAllFieldValuesRegexes(ctx, subSystemName, regexp.MustCompile(gitlabPersonlAccessTokenRegex))

	return ctx
}
