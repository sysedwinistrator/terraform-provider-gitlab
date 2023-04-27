package sdk

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/xanzy/go-gitlab"
)

var _ = registerResource("gitlab_pipeline_schedule_variable", func() *schema.Resource {
	return &schema.Resource{
		Description: `The ` + "`" + `gitlab_pipeline_schedule_variable` + "`" + ` resource allows to manage the lifecycle of a variable for a pipeline schedule.

**Upstream API**: [GitLab REST API docs](https://docs.gitlab.com/ee/api/pipeline_schedules.html#pipeline-schedule-variables)`,

		CreateContext: resourceGitlabPipelineScheduleVariableCreate,
		ReadContext:   resourceGitlabPipelineScheduleVariableRead,
		UpdateContext: resourceGitlabPipelineScheduleVariableUpdate,
		DeleteContext: resourceGitlabPipelineScheduleVariableDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema:        gitlabPipelineScheduleVariableSchema(),
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceGitlabPipelineScheduleVariableResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceGitlabPipelineScheduleVariableStateUpgradeV0,
				Version: 0,
			},
		},
	}
})

func gitlabPipelineScheduleVariableSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"project": {
			Description: "The id of the project to add the schedule to.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"pipeline_schedule_id": {
			Description: "The id of the pipeline schedule.",
			Type:        schema.TypeInt,
			Required:    true,
			ForceNew:    true,
		},
		"key": {
			Description: "Name of the variable.",
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
		},
		"value": {
			Description: "Value of the variable.",
			Type:        schema.TypeString,
			Required:    true,
		},
	}
}

// resourceGitlabPipelineScheduleVariableResourceV0 returns the V0 schema definition.
// From V0-V1 the `id` attribute value format changed from `<pipeline-schedule-variable-id>` to `<project-id>:<pipeline-schedule-variable-id>:<key>`,
// which means that the actual schema definition was not impacted and we can just return the
// V1 schema as V0 schema.
func resourceGitlabPipelineScheduleVariableResourceV0() *schema.Resource {
	return &schema.Resource{Schema: gitlabPipelineScheduleVariableSchema()}
}

// resourceGitlabPipelineScheduleVariableStateUpgradeV0 performs the state migration from V0 to V1.
func resourceGitlabPipelineScheduleVariableStateUpgradeV0(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	project := rawState["project"].(string)
	pipelineScheduleId := rawState["pipeline_schedule_id"].(int)
	key := rawState["key"].(string)

	oldId := rawState["id"].(string)

	tflog.Debug(ctx, "attempting state migration from V0 to V1 - changing the `id` attribute format", map[string]interface{}{"project": project, "pipeline_schedule_id": pipelineScheduleId, "key": key, "v0-id": oldId})
	rawState["id"] = resourceGitlabPipelineScheduleVariableBuildId(project, pipelineScheduleId, key)
	tflog.Debug(ctx, "migrated `id` attribute for V0 to V1", map[string]interface{}{"v0-id": oldId, "v1-id": rawState["id"]})
	return rawState, nil
}

func resourceGitlabPipelineScheduleVariableBuildId(project string, pipelineScheduleId int, key string) string {
	return fmt.Sprintf("%s:%d:%s", project, pipelineScheduleId, key)
}

func resourceGitlabPipelineScheduleVariableParseId(id string) (string, int, string, error) {
	parts := strings.SplitN(id, ":", 3)
	if len(parts) != 3 {
		return "", 0, "", fmt.Errorf("Unexpected ID format (%q). Expected project:pipelineScheduleId:key", id)
	}

	pipelineScheduleId, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", 0, "", err
	}

	return parts[0], pipelineScheduleId, parts[2], nil
}

func resourceGitlabPipelineScheduleVariableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	project := d.Get("project").(string)
	scheduleID := d.Get("pipeline_schedule_id").(int)

	options := &gitlab.CreatePipelineScheduleVariableOptions{
		Key:   gitlab.String(d.Get("key").(string)),
		Value: gitlab.String(d.Get("value").(string)),
	}

	log.Printf("[DEBUG] create gitlab PipelineScheduleVariable %s:%s", *options.Key, *options.Value)

	scheduleVar, _, err := client.PipelineSchedules.CreatePipelineScheduleVariable(project, scheduleID, options, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceGitlabPipelineScheduleVariableBuildId(project, scheduleID, scheduleVar.Key))
	return resourceGitlabPipelineScheduleVariableRead(ctx, d, meta)
}

func resourceGitlabPipelineScheduleVariableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)

	project, scheduleID, pipelineVariableKey, err := resourceGitlabPipelineScheduleVariableParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] read gitlab PipelineSchedule %s/%d", project, scheduleID)

	pipelineSchedule, _, err := client.PipelineSchedules.GetPipelineSchedule(project, scheduleID, gitlab.WithContext(ctx))
	if err != nil {
		return diag.FromErr(err)
	}

	found := false
	for _, pipelineVariable := range pipelineSchedule.Variables {
		if pipelineVariable.Key == pipelineVariableKey {
			d.Set("project", project)
			d.Set("key", pipelineVariable.Key)
			d.Set("value", pipelineVariable.Value)
			d.Set("pipeline_schedule_id", scheduleID)
			found = true
			break
		}
	}
	if !found {
		log.Printf("[DEBUG] pipeline schedule variable not found %s/%d/%s", project, scheduleID, pipelineVariableKey)
		d.SetId("")
	}

	return nil
}

func resourceGitlabPipelineScheduleVariableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	project, scheduleID, variableKey, err := resourceGitlabPipelineScheduleVariableParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if d.HasChange("value") {
		options := &gitlab.EditPipelineScheduleVariableOptions{
			Value: gitlab.String(d.Get("value").(string)),
		}

		log.Printf("[DEBUG] update gitlab PipelineScheduleVariable %s", d.Id())

		_, _, err := client.PipelineSchedules.EditPipelineScheduleVariable(project, scheduleID, variableKey, options, gitlab.WithContext(ctx))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceGitlabPipelineScheduleVariableRead(ctx, d, meta)
}

func resourceGitlabPipelineScheduleVariableDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*gitlab.Client)
	project, scheduleID, variableKey, err := resourceGitlabPipelineScheduleVariableParseId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if _, _, err := client.PipelineSchedules.DeletePipelineScheduleVariable(project, scheduleID, variableKey, gitlab.WithContext(ctx)); err != nil {
		return diag.Errorf("%s failed to delete pipeline schedule variable: %s", d.Id(), err.Error())
	}
	return nil
}
