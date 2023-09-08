package vercel

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vercel/terraform-provider-vercel/client"
)

// Team reflects the state terraform stores internally for a team.
type Team struct {
	ID     types.String `tfsdk:"id"`
	Avatar types.String `tfsdk:"avatar"` // omitempty
	// Billing *Billing `tfsdk:"billing"` // TODO(zchee): implements
	Membership *TeamMember  `tfsdk:"membership"`
	Name       types.String `tfsdk:"name"`
	Slug       types.String `tfsdk:"slug"`

	// PlatformVersion *PlatformVersion `tfsdk:"platformVersion"` // TODO(zchee): implements
	InviteCode          types.String         `tfsdk:"inviteCode"`
	Description         types.String         `tfsdk:"description"`
	StagingPrefix       types.String         `tfsdk:"stagingPrefix"`
	ResourceConfig      *ResourceConfig      `tfsdk:"resourceConfig"`
	RemoteCaching       *enabled             `tfsdk:"remoteCaching"`
	EnabledInvoiceItems *EnabledInvoiceItems `tfsdk:"enabledInvoiceItems"`
	Spaces              *enabled             `tfsdk:"spaces"`
}

var nullTeam = Team{
	// As this is read only, none of these fields are specified - so treat them all as Null
	Avatar:      types.StringNull(),
	Description: types.StringNull(),
}

type ResourceConfig struct {
	ConcurrentBuilds types.Int64 `tfsdk:"concurrentBuilds"`
}

type EnabledInvoiceItems struct {
	ConcurrentBuilds types.Bool `tfsdk:"concurrentBuilds"`
	Monitoring       types.Bool `tfsdk:"monitoring"`
}

type enabled struct {
	Value types.Bool `tfsdk:"enabled"`
}

func enabledToBool(enabled *client.Enabled) types.Bool {
	if enabled == nil || enabled.Value == nil {
		return types.BoolNull()
	}
	return types.BoolPointerValue(enabled.Value)
}

func convertResponseToTeam(response client.TeamResponse, plan Team) Team {
	var ms *TeamMember
	if msResp := response.Membership; msResp != nil {
		ms = &TeamMember{
			Role:      fromStringPointer(msResp.Role),
			Confirmed: fromBoolPointer(msResp.Confirmed),
			UID:       fromStringPointer(msResp.UID),
		}
		if msResp.JoinedFrom != nil {
			ms.JoinedFrom = &TeamJoinedFrom{
				SSOUserID: fromStringPointer(msResp.JoinedFrom.Origin),
			}
		}
	}

	var ii *EnabledInvoiceItems
	if iiResp := response.EnabledInvoiceItems; iiResp != nil {
		ii.ConcurrentBuilds = enabledToBool(iiResp.ConcurrentBuilds)
		ii.Monitoring = enabledToBool(iiResp.Monitoring)
	}

	var rc *ResourceConfig
	if response.ResourceConfig != nil {
		rc.ConcurrentBuilds = types.Int64PointerValue(response.ResourceConfig.ConcurrentBuilds)
	}

	return Team{
		ID:                  types.StringValue(response.ID),
		Avatar:              types.StringPointerValue(response.Avatar),
		Membership:          ms,
		Name:                types.StringPointerValue(response.Name),
		InviteCode:          types.StringPointerValue(response.InviteCode),
		Description:         types.StringPointerValue(response.Description),
		StagingPrefix:       types.StringPointerValue(response.StagingPrefix),
		ResourceConfig:      rc,
		RemoteCaching:       &enabled{Value: enabledToBool(response.RemoteCaching)},
		EnabledInvoiceItems: ii,
		Spaces:              &enabled{Value: enabledToBool(response.RemoteCaching)},
	}
}
