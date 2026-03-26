package gerritsdk

import (
	"context"
	"fmt"

	gogerrit "github.com/andygrunwald/go-gerrit"
)

// ListProjects returns all projects on the Gerrit instance.
func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	opts := &gogerrit.ProjectOptions{}
	projects, _, err := c.inner.Projects.ListProjects(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	result := make([]Project, 0, len(*projects))
	for name, info := range *projects {
		result = append(result, Project{
			ID:          info.ID,
			Name:        name,
			Description: info.Description,
			State:       info.State,
		})
	}
	return result, nil
}

// GetProject returns details about a specific project.
func (c *Client) GetProject(ctx context.Context, name string) (*Project, error) {
	info, _, err := c.inner.Projects.GetProject(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("get project %s: %w", name, err)
	}
	return &Project{
		ID:          info.ID,
		Name:        name,
		Description: info.Description,
		State:       info.State,
	}, nil
}

// ListProjectBranches returns branches for a project.
func (c *Client) ListProjectBranches(ctx context.Context, project string) ([]string, error) {
	opts := &gogerrit.BranchOptions{}
	branches, _, err := c.inner.Projects.ListBranches(ctx, project, opts)
	if err != nil {
		return nil, fmt.Errorf("list branches for %s: %w", project, err)
	}

	result := make([]string, 0, len(*branches))
	for _, b := range *branches {
		result = append(result, b.Ref)
	}
	return result, nil
}
