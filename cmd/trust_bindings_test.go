package cmd

import "testing"

func TestTBCreate_ToInput(t *testing.T) {
	c := tbCreateCmd{
		BindingType:  "custom_ci",
		Identity:     `{"repository":"octocat/hello-world","workflow":"deploy.yml"}`,
		OutputClaims: `{"tier":"prod"}`,
	}
	in, err := c.toInput()
	if err != nil {
		t.Fatalf("toInput: %v", err)
	}
	if in.BindingType != "custom_ci" {
		t.Fatalf("BindingType = %q, want custom_ci", in.BindingType)
	}
	if in.Identity["repository"] != "octocat/hello-world" {
		t.Fatalf("Identity[repository] = %v", in.Identity["repository"])
	}
	if in.OutputClaims["tier"] != "prod" {
		t.Fatalf("OutputClaims[tier] = %v", in.OutputClaims["tier"])
	}
}

func TestTBCreate_ToInput_InvalidIdentityJSON(t *testing.T) {
	c := tbCreateCmd{
		BindingType: "custom_ci",
		Identity:    `{bad json`,
	}
	if _, err := c.toInput(); err == nil {
		t.Fatal("expected error for invalid identity JSON")
	}
}

func TestTBEnableTerraformCloud_ToInput(t *testing.T) {
	c := tbEnableTerraformCloudCmd{
		TFCOrganization: "acme",
		TFCWorkspace:    "prod",
	}
	in := c.toInput()
	if in.BindingType != "terraform_cloud" {
		t.Fatalf("BindingType = %q, want terraform_cloud", in.BindingType)
	}
	if in.Identity["terraform_organization_name"] != "acme" {
		t.Fatalf("terraform_organization_name = %v", in.Identity["terraform_organization_name"])
	}
	if in.Identity["terraform_workspace_name"] != "prod" {
		t.Fatalf("terraform_workspace_name = %v", in.Identity["terraform_workspace_name"])
	}
}

func TestTBEnableGitHubActions_ToInput(t *testing.T) {
	c := tbEnableGitHubActionsCmd{
		Repository: "octocat/hello-world",
		Workflow:   "deploy.yml",
	}
	in := c.toInput()
	if in.BindingType != "github_actions" {
		t.Fatalf("BindingType = %q, want github_actions", in.BindingType)
	}
	if in.Identity["repository"] != "octocat/hello-world" {
		t.Fatalf("repository = %v", in.Identity["repository"])
	}
	if in.Identity["workflow"] != "deploy.yml" {
		t.Fatalf("workflow = %v", in.Identity["workflow"])
	}
}
