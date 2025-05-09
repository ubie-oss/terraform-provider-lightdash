# Lightdash Space and Space Access Notes

This document provides detailed notes on Lightdash spaces and access management, focusing on aspects relevant to the implementation of the Terraform provider. It summarizes key behaviors, API interactions, inheritance rules, and specific considerations for state management.

---

## 1. Space Overview

### 1.1 Space Roles

Lightdash defines the following roles for managing permissions within a space:

- **Full Access (`admin`):** Highest level. Users can manage the space and its contents, including user/group access and editing space details (name, description). They can also adjust a user's inherited project permissions specifically within that space.
- **Can Edit (`editor`):** Users can modify content within the space, such as adding, deleting, or renaming charts and dashboards.
- **Can View (`viewer`):** Minimum access level to see a space and its contents. Users can only view charts and dashboards within the space.

### 1.2 Space Visibility

Spaces have different visibility levels, impacting who can access them by default:

- **Public Access:** Accessible to everyone with access to the project the space belongs to. By default, users inherit space permissions from their project-level permissions.
- **Restricted Access:** A broader category for spaces with limited access. Restricted spaces can function as:

  - **Private:** Only the space creator (initially the admin) has access.
  - **Shared:** Access is limited to organization administrators and specific users or groups explicitly invited or inheriting access.

    To remove a user's access to a space that was previously public, its visibility **must** be changed to Restricted.

### 1.3 Nested Space Behavior

Notes on nested spaces (while official documentation states this is not currently supported, these are observed behaviors):

- Nested spaces inherit the visibility and access controls of their root-level parent space.
- It is impossible to change the visibility of nested spaces independently.
- Space access granted to groups at the root level is inherited by nested spaces.
- It is impossible to create private/shared nested spaces under a public root space, and vice versa.

---

## 2. Managing Spaces (API Operations - v1)

This section outlines key API interactions for managing spaces.

### 2.1 Get Space

The `GetSpace` API call provides detailed information, including how a user's access is granted. The `results.access[]` field indicates:

- `hasDirectAccess`: True if the user was explicitly granted access.
- `inheritedRole`: The role inherited from an upper level (e.g., organization or group).
- `inheritedFrom`: The source of the inherited role.
- `projectRole`: The user's role within the associated project.

### 2.2 Create Space

When a space is created, the creator is automatically assigned the `Full Access` role for that space.

### 2.3 Update Space

The update API is used for both root-level and nested spaces. However, editable fields differ:

- **Root-level spaces:** Multiple fields (e.g., name, visibility, access grants) can be updated.
- **Nested spaces:** Only the `name` field can be updated via the v1 API at this time.

### 2.4 Move Space

Moving a space to a different parent has a dedicated API endpoint. When a space is moved, it **inherits** the permissions of the new parent space, overriding its previous access configuration.

### 2.5 Delete Space

Deleting a space also deletes all its descendant spaces. The Terraform resource `lightdash_space` should not allow deleting a space if it has descendant spaces to prevent accidental data loss and align with typical resource lifecycle management.

---

## 3. Space Access and Inheritance

Space access is determined by considering several factors, with explicit assignments taking precedence.

### 3.1 Inheritance Mechanisms

Access can be inherited through:

- **Direct Assignment:** Explicitly inviting users or groups with a specific role.
- **Group Membership:** Users inherit permissions from groups granted access. The highest role from all relevant groups applies.
- **Organization Admin:** Organization administrators automatically have unrevokable `Full Access` to all spaces.
- **Project Permissions (Public Spaces):** In public spaces, users inherit space permissions based on their project role by default.
- **Ancestor Spaces (Nested Spaces):** Nested spaces inherit the access configuration of their root-level parent.

### 3.2 Overrides

Explicitly assigned user access to a space always overrides any inherited access from groups or project roles for that specific user.

### 3.3 Limitation: Minimum Members

A space must always have at least one member with access. It is impossible to completely revoke all access from a space.

---

## 4. Terraform Provider Implementation Notes

This section details specific considerations for managing space access within the Terraform provider.

### 4.1 State Management for Member Access (`access` vs `access_all`)

Similar to how AWS providers handle resource tags, we should distinguish between the desired state defined in the Terraform configuration and the effective state returned by the Lightdash API. This helps manage situations where Lightdash automatically adds members (like the creator or org admins) that are not explicitly defined in the Terraform plan.

- **`access` Attribute:** This attribute in the Terraform resource schema should reflect the explicit member access grants defined by the practitioner in the `.tf` configuration. It should be written only with the values from `req.Plan` during Create/Update operations.
- **`access_all` Attribute:** This attribute should store the complete list of members with access to the space, as returned by the Lightdash API (including the creator, org admins, and any other inherited access). It should be written with the full list returned by Lightdash's `GetSpace` API call during Create/Update/Read operations.

This pattern allows practitioners to define their desired state using `access`, while `access_all` tracks the effective state. Terraform will detect drift if the effective state (e.g., the creator's access) is modified outside of Terraform, as `access_all` will change, even if the `access` attribute defined in the configuration remains the same.

This approach addresses the challenge of Lightdash automatically including the creator as a space admin, ensuring the provider can manage explicit grants without attempting to remove automatically added members.

### 4.2 State Management for Group Access (`group_access` vs `group_access_all`)

The same principle applies to group access grants:

- **`group_access` Attribute:** Should store the explicit group access grants defined by the practitioner in the `.tf` configuration.
- **`group_access_all` Attribute:** Should store the complete list of groups with access, as returned by the Lightdash API (including any inherited group access).
