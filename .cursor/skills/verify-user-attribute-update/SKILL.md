---
name: verify-user-attribute-update
description: Verifies the user values of lightdash_user_attribute during Update with a TLA+ model and a code conformance check. Use when changing internal/lightdash/services/user_attributes.go (Update, GetByUUID, keepUserValues, toUserAttributeV1Request), internal/lightdash/api/v1/update_user_attribute_v1.go, or the Update method of internal/provider/resource_lightdash_user_attribute.go, or when asked whether a user attribute update can lose users.
---

# Verify the user attribute update

## What is verified

Property: a user value set in the Lightdash UI survives a Terraform update of the same attribute when no edit happens during the update.

Harm: Lightdash `PUT /api/v1/org/attributes/{uuid}` replaces the whole attribute, including its users. If `Update` writes the attribute without the users it has just read, every user value is erased on each apply.

The model covers `Update` as two steps:

- Get: `UserAttributesService.Update` calls `GetByUUID`, which lists the attributes and reads the users.
- Put: `keepUserValues` copies the read users into the request, then `apiv1.UpdateUserAttributeV1` sends it.

Between the two steps, a user can edit the value in the Lightdash UI (UIEdit).

Model: `dev/formal/user-attribute/UserAttributeUpdate.tla`. Tool: TLA+ with TLC. A model checker explores every order of the Get, Put, and UI-edit steps. A plain test would only cover the orders someone wrote down.

## Coverage record

| Model piece              | Code it describes                                                                                                                               |
| ------------------------ | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| Get                      | `internal/lightdash/services/user_attributes.go` `Update`, the `GetByUUID` call                                                                 |
| Put, `KeepUsers = TRUE`  | `keepUserValues` (`desired.Users = current.Users`) and `apiv1.UpdateUserAttributeV1` in `internal/lightdash/api/v1/update_user_attribute_v1.go` |
| Put, `KeepUsers = FALSE` | A regression in which the write sends no users. Used only as a self-test.                                                                       |
| UIEdit                   | A user changes the value in the Lightdash UI, outside Terraform                                                                                 |
| Resource Update          | `internal/provider/resource_lightdash_user_attribute.go` `Update`, which calls `r.service.Update`                                               |

Sync commit: `e731689`. The verification describes the code at this commit. This is the commit where `main` contained the current Update path (merge of #514).

When the covered code changes after this commit, the verification may no longer describe it. Run the check below first.

## Run

```bash
dev/formal/user-attribute/check.sh
```

Needs Java 11 or later and `rg`. The script downloads `tla2tools.jar` 1.7.4 into `~/.cache` and checks its SHA-256. To use a local copy, set `TLA2TOOLS=/path/to/tla2tools.jar`.

Exit codes:

- `0`: the model holds for the code. Not expected today, see the finding below.
- `1`: a finding. The check ran, and TLC found an interleaving in which a user value is lost. The counterexample is printed.
- `2`: drift or TLC failure. Either the code no longer matches the model, or a model config did not behave as expected. Read the `DRIFT:` line.

The three TLC configs and what each must show:

| Config                   | Expected result    | Meaning                                                                                           |
| ------------------------ | ------------------ | ------------------------------------------------------------------------------------------------- |
| `sequential.cfg`         | PASS               | No UI edit during an apply: values survive. If it fails, the model or the invariant is broken.    |
| `keep-users-removed.cfg` | VIOLATED           | A regression that drops the users is caught. If it passes, the invariant cannot see a lost value. |
| `race.cfg`               | VIOLATED (finding) | A UI edit between Get and Put is lost. This is the current behavior in `main`.                    |

## Read a result

Read every TLC counterexample with the pstack `/statechart` skill. It turns each violated invariant's shortest path into plain steps, so you can tell a real race in the code from a modeling error.

For the current finding, the path is four states long:

1. Get reads the users `us`.
2. A user edits the value to `eu` in the Lightdash UI.
3. Put sends the users that Get read (`us`).
4. The stored value is `us`, and the UI edit is gone.

Decide before you act:

- If the path matches the code (`Update` reads, then writes with the read users), it is a real finding. Report it. Do not weaken `race.cfg`, the invariant, or `EditsDuringApply` to make it pass.
- If the path does not match the code, the model is wrong. Fix the model, then rerun the check.

## Update the verification

Change the verification only when the change to the code is intended, or when the model is wrong.

1. Read the diff of the covered files since `e731689`.
2. Update `UserAttributeUpdate.tla`, the configs, `check.sh` (the `rg` patterns and the function names), and the coverage table above.
3. Run `dev/formal/user-attribute/check.sh`. Expect `sequential.cfg` PASS, `keep-users-removed.cfg` VIOLATED, and `race.cfg` VIOLATED. Exit 1.
4. Do the scratch-break check below.
5. Move the sync commit to the new commit only after you confirm that the model still describes the code there. Do not move it because the check is green.

## Scratch-break check

After any change to the check, confirm that it still catches the break it guards against:

1. In `internal/lightdash/services/user_attributes.go`, delete the line `attribute = keepUserValues(attribute, *current)` in `Update`.
2. Run `dev/formal/user-attribute/check.sh`. Expect exit `2` with `DRIFT: Update no longer copies the users that it read into the write`.
3. Restore the file with `git checkout -- internal/lightdash/services/user_attributes.go`.
4. Run the check again. Expect exit `1` with only the race finding.

## Not covered

- Read, import, and drift detection. Read removes the resource from state only when Lightdash reports the attribute missing, and the model does not cover that.
- Create and Delete. They have no read-then-write step in this model.
- Live Lightdash behavior, such as whether the API really replaces the whole attribute. The model takes that as given from the code and the API client.
- Other resources and the `lightdash_user_attribute` data source.
- Several concurrent Terraform applies, and more than one UI edit per run beyond the `MaxEdits` bound (`2` in `sequential.cfg`, `1` in `race.cfg`).
- Several user values per attribute. The model tracks one value for one attribute.
