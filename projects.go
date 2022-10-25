package tfe

import (
	"context"
	"fmt"
	"log"
	"net/url"
)

// Compile-time proof of interface implementation.
var _ Projects = (*projects)(nil)

// Projects describes all the project related methods that the Terraform
// Enterprise API supports
//
// TFE API docs: (TODO: ADD DOCS URL)
type Projects interface {
	// List all projects in the given organization
	//List(ctx context.Context, organization string, options *ProjectListOptions) (*ProjectList, error)

	// Create a new project.
	Create(ctx context.Context, organization string, options ProjectCreateOptions) (*Project, error)

	// Read a project by its ID.
	Read(ctx context.Context, ProjectID string) (*Project, error)

	// Update a project.
	Update(ctx context.Context, ProjectID string, options ProjectUpdateOptions) (*Project, error)

	//Delete a project.
	Delete(ctx context.Context, ProjectID string) error
}

// projects implements Projects
type projects struct {
	client *Client
}

// ProjectList represents a list of projects
type ProjectList struct {
	*Pagination
	Items []*Project
}

// Project represents a Terraform Enterprise project
type Project struct {
	ID   string `jsonapi:"primary,projects"`
	Name string `jsonapi:"attr,name"`

	// Relations
	Organization *Organization `jsonapi:"relation,organization"`
}

//// ProjectListOptions represents the options for listing projects
//type ProjectListOptions struct {
//	ListOptions
//
//	// Add more list options here
//}

// ProjectCreateOptions represents the options for creating a project
type ProjectCreateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,projects"`

	// Required: A name to identify the agent pool.
	Name *string `jsonapi:"attr,name"`
}

// ProjectUpdateOptions represents the options for updating a project
type ProjectUpdateOptions struct {
	// Type is a public field utilized by JSON:API to
	// set the resource type via the field tag.
	// It is not a user-defined value and does not need to be set.
	// https://jsonapi.org/format/#crud-creating
	Type string `jsonapi:"primary,projects"`

	// Required: A name to identify the agent pool.
	Name *string `jsonapi:"attr,name"`
}

// List all projects.
//func List(ctx context.Context, options *ProjectListOptions) (*ProjectList, error) {
//	panic("not yet implemented")
//}

// Create a project with the given options
func (s *projects) Create(ctx context.Context, organization string, options ProjectCreateOptions) (*Project, error) {
	if !validStringID(&organization) {
		return nil, ErrInvalidOrg
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("organizations/%s/projects", url.QueryEscape(organization))
	req, err := s.client.NewRequest("POST", u, &options)
	if err != nil {
		return nil, err
	}

	p := &Project{}
	err = req.Do(ctx, p)
	log.Println(ctx)
	log.Println(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Read a single project by its ID.
func (s *projects) Read(ctx context.Context, ProjectID string) (*Project, error) {
	if !validStringID(&ProjectID) {
		return nil, ErrInvalidProjectID
	}

	u := fmt.Sprintf("projects/%s", url.QueryEscape(ProjectID))
	req, err := s.client.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	p := &Project{}
	err = req.Do(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Update a project by its ID
func (s *projects) Update(ctx context.Context, ProjectID string, options ProjectUpdateOptions) (*Project, error) {
	if !validStringID(&ProjectID) {
		return nil, ErrInvalidProjectID
	}

	if err := options.valid(); err != nil {
		return nil, err
	}

	u := fmt.Sprintf("projects/%s", url.QueryEscape(ProjectID))
	req, err := s.client.NewRequest("PATCH", u, &options)
	if err != nil {
		return nil, err
	}

	p := &Project{}
	err = req.Do(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Delete a project by its ID
func (s *projects) Delete(ctx context.Context, ProjectID string) error {
	if !validStringID(&ProjectID) {
		return ErrInvalidProjectID
	}

	u := fmt.Sprintf("projects/%s", url.QueryEscape(ProjectID))
	req, err := s.client.NewRequest("DELETE", u, nil)
	if err != nil {
		return err
	}

	return req.Do(ctx, nil)
}

func (o ProjectCreateOptions) valid() error {
	if !validString(o.Name) {
		return ErrRequiredName
	}
	if !validStringID(o.Name) {
		return ErrInvalidName
	}
	return nil
}

func (o ProjectUpdateOptions) valid() error {
	if o.Name != nil && !validStringID(o.Name) {
		return ErrInvalidName
	}
	return nil
}
