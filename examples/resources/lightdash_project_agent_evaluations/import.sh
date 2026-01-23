#!/bin/sh

# The resource "lightdash_project_agent_evaluations" can be imported using the organization, project, agent, and evaluation UUIDs.
# The ID is in the format "organizations/<organization_uuid>/projects/<project_uuid>/agents/<agent_uuid>/evaluations/<evaluation_uuid>".

organization_uuid="xxxx-xxxx-xxxx"
project_uuid="xxxx-xxxx-xxxx"
agent_uuid="xxxx-xxxx-xxxx"
evaluation_uuid="xxxx-xxxx-xxxx"

terraform import lightdash_project_agent_evaluations.test "organizations/${organization_uuid}/projects/${project_uuid}/agents/${agent_uuid}/evaluations/${evaluation_uuid}"
