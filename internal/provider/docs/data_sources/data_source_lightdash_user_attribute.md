Reads a Lightdash user attribute of the organization by its name, including its `attribute_defaults` and the values assigned to groups. Use it to refer to a user attribute that this configuration does not manage.

The name must match exactly one user attribute. The data source fails when no user attribute has the name, and when more than one has it. It does not return values assigned to individual users.
