---
name: verify-org-role-member-lifecycle
description: Check the TLA+ model of the lightdash_organization_role_member lifecycle (Create, Update, Delete and the org role catalog lookup) against the code. Use when a change touches internal/provider/resource_lightdash_organization_role_member.go, AssignOrgUserRole or ResolveRoleID in internal/lightdash/services/roles.go, GetOrganizationRolesV2, AssignOrganizationRoleToUserV2, or the org role catalog behavior, or when asked to re-verify the role member lifecycle or issue #461.
---

# Verify the organization role member lifecycle

## What is verified

Property: **every built-in organization role can be assigned by Create, Update, and Delete.**
Harm: a valid role such as `member` fails with `role "member" not found`, so apply fails
and `terraform destroy` cannot reset the user (issue #461).

The model is `dev/formal/org-role-member/OrgRoleMember.tla`. It has three invariants:

- `NoRoleNotFound`: the last operation succeeded.
- `StateMatchesAssignment`: while the resource is managed, Lightdash and Terraform state agree.
- `UnmanagedIsDefault`: once the resource is gone, the user is back on `member`.

The model covers Terraform-only changes. It does not cover Read, drift, out-of-band edits, or the
`GET /api/v2/orgs/{org}/roles` response beyond the catalog contents named in `live.cfg`.

## Coverage record

| Model part                           | Code                                                                                                                                   |
| ------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------- |
| `Create`                             | `Create` in `internal/provider/resource_lightdash_organization_role_member.go`, calls `AssignOrgUserRole` with the planned role        |
| `Update`                             | `Update` in the same file, calls `AssignOrgUserRole` with the planned role                                                             |
| `Delete`                             | `Delete` in the same file, calls `AssignOrgUserRole` with `models.ORGANIZATION_MEMBER_ROLE`                                            |
| `CanAssign` with `UseCatalog = TRUE` | `AssignOrgUserRole` then `ResolveRoleID` in `internal/lightdash/services/roles.go`, which uses `GetRoles` and `resolveRoleIDFromRoles` |
| `Catalog`                            | `GetOrganizationRolesV2` in `internal/lightdash/api/v2/get_organization_roles_v2.go`                                                   |
| Assignment POST                      | `AssignOrganizationRoleToUserV2` in `internal/lightdash/api/v2/assign_organization_role_to_user_v2.go`                                 |

Catalog contents in `live.cfg` come from issue #461: project-level slugs and custom roles, no
org-level `member`. `fixture.cfg` uses the catalog that the unit and contract tests assume,
which includes `member`.

**Sync commit: `f45af77`.** At this commit, `AssignOrgUserRole` sends built-in org role slugs
without a catalog lookup, which matches `slug.cfg`. The pre-fix commit `991b386` called
`ResolveRoleID` and matched `live.cfg`. Move this commit forward only after
`dev/formal/org-role-member/check.sh` exits 0 or 1 with the same findings on the new commit, and
you have confirmed that the table above still describes the code.

## Run

```sh
dev/formal/org-role-member/check.sh
```

The script:

1. Checks conformance. Create, Update, and Delete still assign roles through `AssignOrgUserRole`,
   and Delete still resets to `member`. Exit 2 means drift.
2. Runs TLC on `fixture.cfg`. It must pass. Exit 2 means the model's invariants no longer hold
   under the test catalog.
3. Detects whether `AssignOrgUserRole` resolves names through `ResolveRoleID`. If yes, it runs
   `live.cfg`. If no, it runs `slug.cfg`.

Exit 0 means the model holds for the code. Exit 1 means a finding: the code breaks
`NoRoleNotFound`, and TLC's counterexample is printed. Exit 2 means drift or a failed TLC run.

The script needs Java 11 or later and `rg`. It downloads TLC 1.7.4 and checks its SHA-256,
unless `TLA2TOOLS` points at a jar.

## Read a result

- Show each TLC counterexample to the reader with `/statechart`. It turns a bug path into a
  chart of states and steps, and it names the broken rule in plain words. Do not copy
  `/statechart` commands, options, or paths into this skill or its scripts. Look them up in
  that skill when you need them.
- A counterexample is a finding only after you replay it on the code. For `NoRoleNotFound`,
  replay: Create the resource with `member` in an org whose catalog lacks `member`, then check
  that the provider reports `role "member" not found`.
- If conformance fails (exit 2), the code changed. Decide whether the change is intended.
  If the change is intended, update the model and its config. If it is not, the code is wrong.
- If `fixture.cfg` fails, the invariants changed in the model. Find the edit in the spec, and do
  not weaken an invariant to make it pass.

## Update the verification

After an intended change to the covered code:

1. Update the coverage record table with the new function names or file paths.
2. Update `OrgRoleMember.tla` and the matching config so that each action matches the code.
3. Run `check.sh`. Confirm it exits 0 or 1 with the expected findings.
4. Move the sync commit forward only after step 3 confirms that the verification describes the new code.

## Scratch-break check

To confirm the script catches a break, change `Delete` in
`internal/provider/resource_lightdash_organization_role_member.go` to use
`models.ORGANIZATION_VIEWER_ROLE`. `check.sh` must exit 2 with
`DRIFT: Delete no longer resets the user to the default role member`. Then restore the file with
`git checkout -- internal/provider/resource_lightdash_organization_role_member.go`.

## Not covered

- Read, drift detection, and import.
- Concurrency, and the process-wide role catalog cache in `RoleService`.
- Behavior of the live API beyond the catalog contents in `live.cfg`. Only a live check can
  confirm whether `POST /api/v2/orgs/{org}/roles/assignments/user/{user}` accepts the slug `member`.
- Other resources, including `lightdash_project_role_member`.
