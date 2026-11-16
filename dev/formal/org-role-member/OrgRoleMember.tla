---- MODULE OrgRoleMember ----
(*
Lifecycle of the Terraform resource lightdash_organization_role_member.

Create, Update, and Delete (internal/provider/resource_lightdash_organization_role_member.go)
each call RoleService.AssignOrgUserRole (internal/lightdash/services/roles.go).
With UseCatalog = TRUE, that call first looks the role name up in the organization
role catalog (GET /api/v2/orgs/{org}/roles) and fails with "role not found" when the
name is missing. Delete always assigns the default role "member".

Scope: Terraform-only changes. No out-of-band edits and no Read/drift.
*)
EXTENDS TLC

CONSTANTS
  Roles,      \* built-in organization roles that the resource accepts
  Catalog,    \* role slugs that the organization role catalog returns
  UseCatalog  \* TRUE: names resolve through Catalog; FALSE: the slug is sent as-is

ASSUME "member" \in Roles
ASSUME Catalog \subseteq Roles
ASSUME UseCatalog \in BOOLEAN

VARIABLES
  assigned,   \* the user's organization role in Lightdash
  managed,    \* the resource is in Terraform state
  stateRole,  \* the role that Terraform state records
  lastOk      \* TRUE when the last operation succeeded

vars == <<assigned, managed, stateRole, lastOk>>

Default == "member"

\* Mirrors RoleService.AssignOrgUserRole: with the catalog, the name must be listed.
CanAssign(role) == UseCatalog => role \in Catalog

Init ==
  /\ assigned = Default
  /\ managed = FALSE
  /\ stateRole = Default
  /\ lastOk = TRUE

\* Terraform Create: resource_lightdash_organization_role_member.go Create.
Create(role) ==
  /\ ~managed
  /\ IF CanAssign(role)
     THEN /\ assigned' = role
          /\ managed' = TRUE
          /\ stateRole' = role
          /\ lastOk' = TRUE
     ELSE /\ UNCHANGED <<assigned, managed, stateRole>>
          /\ lastOk' = FALSE

\* Terraform Update: resource_lightdash_organization_role_member.go Update.
Update(role) ==
  /\ managed
  /\ IF CanAssign(role)
     THEN /\ assigned' = role
          /\ stateRole' = role
          /\ UNCHANGED managed
          /\ lastOk' = TRUE
     ELSE /\ UNCHANGED <<assigned, managed, stateRole>>
          /\ lastOk' = FALSE

\* Terraform Delete: resource_lightdash_organization_role_member.go Delete.
\* A failed assignment keeps the resource in Terraform state.
Delete ==
  /\ managed
  /\ IF CanAssign(Default)
     THEN /\ assigned' = Default
          /\ managed' = FALSE
          /\ UNCHANGED stateRole
          /\ lastOk' = TRUE
     ELSE /\ UNCHANGED <<assigned, managed, stateRole>>
          /\ lastOk' = FALSE

CreateAny == \E role \in Roles : Create(role)
UpdateAny == \E role \in Roles : Update(role)
DeleteAny == Delete

Next == CreateAny \/ UpdateAny \/ DeleteAny

Spec == Init /\ [][Next]_vars

\* Every built-in role can be assigned by Create, Update, and Delete.
NoRoleNotFound == lastOk

\* While the resource is managed, Lightdash and Terraform state agree.
StateMatchesAssignment == managed => assigned = stateRole

\* Once the resource is gone, the user is back on the default role.
UnmanagedIsDefault == ~managed => assigned = Default
====
