# Lightdash API v1 → v2 Migration Strategies

This document summarizes known deprecated v1 endpoints and their v2 replacements, for use when auditing the provider or planning migrations.

## Source of truth

- **OpenAPI spec**: `https://raw.githubusercontent.com/lightdash/lightdash/refs/heads/main/packages/backend/src/generated/swagger.json`
- **Docs**: [Lightdash API Reference](https://docs.lightdash.com/api-reference/v1/introduction)

Deprecation is indicated by `deprecated: true` on the operation in the spec.

---

## Project access (user)

| v1 (deprecated)                                           | v2 replacement                                                        | Provider status                                                                                                        |
| --------------------------------------------------------- | --------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `PATCH /api/v1/projects/{projectUuid}/access/{userUuid}`  | `POST /api/v2/projects/{projectId}/roles/assignments/user/{userId}`   | **Migrated** — `lightdash_project_role_member` Update/Delete/Read via `RoleService`; Create still uses v1 invite path. |
| `DELETE /api/v1/projects/{projectUuid}/access/{userUuid}` | `DELETE /api/v2/projects/{projectId}/roles/assignments/user/{userId}` | **Migrated** (see above)                                                                                               |
| `POST /api/v1/projects/{projectUuid}/access`              | `POST /api/v2/projects/{projectId}/roles/assignments/user/{userId}`   | **Partial** — Create uses v1 `GrantProjectAccessToUserV1` (email invite); Update uses v2.                              |

**Migration note**: Resolve role name (e.g. `viewer`, `admin`) to the organization’s `roleId` via `GET /api/v2/orgs/{orgUuid}/roles` (or equivalent) before calling the v2 assignment API. **Note**: v2 assignments use UUIDs for roles, whereas v1 often used strings.

---

## Role UUID Mapping Strategy

When migrating from v1 role strings to v2 role IDs:

1. **Fetch Roles**: Call `GET /api/v2/orgs/{orgUuid}/roles` to get the list of available roles for the organization.
2. **Lookup**: Match the v1 role string (e.g., `developer`) against the `name` or `slug` in the v2 response to find the corresponding `uuid`.
3. **Cache**: Agents should cache this mapping during the migration session to avoid redundant API calls.

---

## Project access (group)

| v1 (deprecated)                                            | v2 replacement                                                                     | Provider status                                                                                                                                                    |
| ---------------------------------------------------------- | ---------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `PUT /api/v1/groups/{groupUuid}/projects/{projectUuid}`    | `POST /api/v2/projects/{projectId}/roles/assignments/group/{groupId}`              | **Migrated** — `lightdash_project_role_group` via `RoleService`.                                                                                                   |
| `PATCH /api/v1/groups/{groupUuid}/projects/{projectUuid}`  | `PATCH /api/v2/projects/{projectId}/roles/assignments/group/{groupId}`             | **Migrated** (see above)                                                                                                                                           |
| `DELETE /api/v1/groups/{groupUuid}/projects/{projectUuid}` | `DELETE /api/v2/projects/{projectId}/roles/assignments/group/{groupId}`            | **Migrated** (see above)                                                                                                                                           |
| `GET /api/v1/projects/{projectUuid}/groupAccesses`         | `GET /api/v2/projects/{projectId}/roles/assignments` (filter `assigneeType=group`) | **Migrated** — `lightdash_project_group_accesses` data source via `RoleService.ListProjectGroupAssignments`. v1 read is not deprecated but unused by the provider. |

**Migration note**: Request body for v2 uses `roleId`; path uses `projectId` and `groupId` (same UUIDs as v1 `projectUuid` / `groupUuid`).

---

## Organization member role

| v1 (deprecated)                      | v2 replacement                                                | Provider status                                                                                                |
| ------------------------------------ | ------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `PATCH /api/v1/org/users/{userUuid}` | `POST /api/v2/orgs/{orgUuid}/roles/assignments/user/{userId}` | **Migrated** — `lightdash_organization_role_member` via `RoleService`. Email on Read still from v1 member API. |

**Migration note**: Organization-level roles are now assignment-based; resolve role name to `roleId` via organization roles API if needed.

---

## Endpoints not deprecated (as of last audit)

The following are still current and do not require migration for deprecation:

- **Spaces**: `GET/POST /api/v1/projects/.../spaces`, `GET/PUT/DELETE .../spaces/{spaceUuid}`, and space share endpoints.
- **Groups (CRUD)**: `GET/POST /api/v1/org/groups`, `GET/PATCH/DELETE /api/v1/groups/{groupUuid}`, and group members.
- **Organization**: `GET /api/v1/org`, `GET /api/v1/org/users`, `GET /api/v1/org/projects`. Member directory listing stays on v1; `ProjectService.GetProjectMembers` intentionally unchanged.
- **Projects**: `GET /api/v1/projects/{projectUuid}`, `PATCH .../schedulerSettings`.
- **AI agents**: All `/api/v1/projects/.../aiAgents` and related evaluation endpoints.
- **Content**: `POST /api/v2/content/{projectUuid}/move` (v2, not deprecated).

---

## Running an audit

1. Use `grep` or `rg` to scan `internal/lightdash/api/` for path patterns (e.g., `"/api/v1/"`).
2. Cross-reference results with the [OpenAPI spec](https://raw.githubusercontent.com/lightdash/lightdash/refs/heads/main/packages/backend/src/generated/swagger.json) to identify `deprecated: true` flags.
3. For any deprecated endpoint, use the `@.claude/skills/research-lightdash-api` skill to confirm live status and find replacements.
4. Draft a migration plan using the mapping tables in this document.
