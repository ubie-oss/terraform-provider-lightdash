// Copyright 2023 Ubie, inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api"
	apiv1 "github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/api/v1"
	"github.com/ubie-oss/terraform-provider-lightdash/internal/lightdash/models"
)

var (
	// ErrUserAttributeNotFound means the organization has no user attribute with that UUID or name.
	ErrUserAttributeNotFound = errors.New("user attribute not found")
	// ErrUserAttributeAmbiguous means more than one user attribute has the name.
	ErrUserAttributeAmbiguous = errors.New("more than one user attribute has the name")
	// ErrUserAttributeOrganizationMismatch means the attribute belongs to another organization.
	ErrUserAttributeOrganizationMismatch = errors.New("user attribute belongs to another organization")
)

type UserAttributesService struct {
	client *api.Client
}

func NewUserAttributesService(client *api.Client) *UserAttributesService {
	return &UserAttributesService{client: client}
}

func (s *UserAttributesService) List(ctx context.Context) ([]models.UserAttribute, error) {
	attributes, err := apiv1.GetUserAttributesV1(s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to list user attributes: %w", err)
	}
	result := make([]models.UserAttribute, 0, len(attributes))
	for _, attribute := range attributes {
		result = append(result, toUserAttributeModel(attribute))
	}
	return result, nil
}

func (s *UserAttributesService) GetByUUID(ctx context.Context, organizationUUID string, userAttributeUUID string) (*models.UserAttribute, error) {
	attributes, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	found := findUserAttributeByUUID(attributes, userAttributeUUID)
	if found == nil {
		return nil, fmt.Errorf("user attribute %s: %w", userAttributeUUID, ErrUserAttributeNotFound)
	}
	if organizationUUID != "" && found.OrganizationUUID != organizationUUID {
		return nil, fmt.Errorf("user attribute %s is in organization %s: %w", userAttributeUUID, found.OrganizationUUID, ErrUserAttributeOrganizationMismatch)
	}
	return found, nil
}

// GetByName returns the one attribute with this exact name.
func (s *UserAttributesService) GetByName(ctx context.Context, name string) (*models.UserAttribute, error) {
	attributes, err := s.List(ctx)
	if err != nil {
		return nil, err
	}
	matches := findUserAttributesByName(attributes, name)
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("user attribute %q: %w", name, ErrUserAttributeNotFound)
	case 1:
		return &matches[0], nil
	default:
		return nil, fmt.Errorf("user attribute %q: %w", name, ErrUserAttributeAmbiguous)
	}
}

func (s *UserAttributesService) Create(ctx context.Context, attribute models.UserAttribute) (*models.UserAttribute, error) {
	created, err := apiv1.CreateUserAttributeV1(s.client, toUserAttributeV1Request(attribute))
	if err != nil {
		return nil, fmt.Errorf("failed to create user attribute %q: %w", attribute.Name, err)
	}
	model := toUserAttributeModel(*created)
	return &model, nil
}

// Update replaces the attribute and keeps the user values that are already stored.
// The API replaces the whole attribute, so omitting users would delete values assigned in the Lightdash UI.
func (s *UserAttributesService) Update(ctx context.Context, organizationUUID string, attribute models.UserAttribute) (*models.UserAttribute, error) {
	current, err := s.GetByUUID(ctx, organizationUUID, attribute.UserAttributeUUID)
	if err != nil {
		return nil, err
	}
	attribute = keepUserValues(attribute, *current)
	updated, err := apiv1.UpdateUserAttributeV1(s.client, attribute.UserAttributeUUID, toUserAttributeV1Request(attribute))
	if err != nil {
		return nil, fmt.Errorf("failed to update user attribute %s: %w", attribute.UserAttributeUUID, err)
	}
	model := toUserAttributeModel(*updated)
	return &model, nil
}

// Delete removes the attribute. It succeeds when the attribute is already gone.
func (s *UserAttributesService) Delete(ctx context.Context, organizationUUID string, userAttributeUUID string) error {
	current, err := s.GetByUUID(ctx, organizationUUID, userAttributeUUID)
	if err != nil {
		if errors.Is(err, ErrUserAttributeNotFound) {
			return nil
		}
		return err
	}
	if err := apiv1.DeleteUserAttributeV1(s.client, current.UserAttributeUUID); err != nil {
		return fmt.Errorf("failed to delete user attribute %s: %w", userAttributeUUID, err)
	}
	return nil
}

func keepUserValues(desired models.UserAttribute, current models.UserAttribute) models.UserAttribute {
	desired.Users = current.Users
	return desired
}

func findUserAttributesByName(attributes []models.UserAttribute, name string) []models.UserAttribute {
	matches := make([]models.UserAttribute, 0)
	for _, attribute := range attributes {
		if attribute.Name == name {
			matches = append(matches, attribute)
		}
	}
	return matches
}

func findUserAttributeByUUID(attributes []models.UserAttribute, userAttributeUUID string) *models.UserAttribute {
	for i := range attributes {
		if attributes[i].UserAttributeUUID == userAttributeUUID {
			return &attributes[i]
		}
	}
	return nil
}

func toUserAttributeModel(attribute apiv1.UserAttributeV1) models.UserAttribute {
	return models.UserAttribute{
		OrganizationUUID:  attribute.OrganizationUUID,
		UserAttributeUUID: attribute.UUID,
		Name:              attribute.Name,
		Description:       attribute.Description,
		AttributeDefaults: attribute.AttributeDefaults,
		Groups:            groupValues(attribute.Groups),
		Users:             userValues(attribute.Users),
		CreatedAt:         attribute.CreatedAt,
	}
}

func groupValues(groups []apiv1.UserAttributeV1Group) []models.UserAttributeValue {
	values := make([]models.UserAttributeValue, 0, len(groups))
	for _, group := range groups {
		values = append(values, models.UserAttributeValue{
			UUID:   group.GroupUUID,
			Values: canonicalValues(group.Values, group.Value),
		})
	}
	return values
}

func userValues(users []apiv1.UserAttributeV1User) []models.UserAttributeValue {
	values := make([]models.UserAttributeValue, 0, len(users))
	for _, user := range users {
		values = append(values, models.UserAttributeValue{
			UUID:   user.UserUUID,
			Values: canonicalValues(user.Values, user.Value),
		})
	}
	return values
}

// canonicalValues prefers the current `values` array. It falls back to the deprecated single `value`.
func canonicalValues(values []string, deprecatedValue string) []string {
	if len(values) > 0 {
		return values
	}
	if strings.TrimSpace(deprecatedValue) == "" {
		return []string{}
	}
	return []string{deprecatedValue}
}

func toUserAttributeV1Request(attribute models.UserAttribute) apiv1.CreateUserAttributeV1Request {
	var defaults *[]string
	if attribute.AttributeDefaults != nil {
		copied := append([]string{}, attribute.AttributeDefaults...)
		defaults = &copied
	}
	return apiv1.CreateUserAttributeV1Request{
		Name:              attribute.Name,
		Description:       attribute.Description,
		AttributeDefaults: defaults,
		Groups:            toGroupRequests(attribute.Groups),
		Users:             toUserRequests(attribute.Users),
	}
}

func toGroupRequests(values []models.UserAttributeValue) []apiv1.CreateUserAttributeGroupV1 {
	groups := make([]apiv1.CreateUserAttributeGroupV1, 0, len(values))
	for _, value := range values {
		groups = append(groups, apiv1.CreateUserAttributeGroupV1{
			GroupUUID: value.UUID,
			Values:    append([]string{}, value.Values...),
		})
	}
	return groups
}

func toUserRequests(values []models.UserAttributeValue) []apiv1.CreateUserAttributeUserV1 {
	users := make([]apiv1.CreateUserAttributeUserV1, 0, len(values))
	for _, value := range values {
		users = append(users, apiv1.CreateUserAttributeUserV1{
			UserUUID: value.UUID,
			Values:   append([]string{}, value.Values...),
		})
	}
	return users
}
