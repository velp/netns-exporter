package projects

import (
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
)

// GetProjectsHierarchy populates projects mapping hierarchy from the provided
// basic projects slice and flattened mapping of the top-level projects:
//
// map[string][]string{
//   basicProjectID: topLevelProjectName,
// }
//
// Projects that don't have associated top-level project names won't be added to
// the hierarchy.
func GetProjectsHierarchy(basic []projects.Project, topLevel map[string]string) map[string]string {
	hierarchy := make(map[string]string, len(basic))

	for _, project := range basic {
		if _, ok := hierarchy[project.ID]; ok {
			// We already added this project into hierarchy.
			continue
		}
		topLevelProjectName, ok := topLevel[project.ParentID]
		if !ok {
			// There is no associated top-level project for the basic project's
			// parent_id field.
			continue
		}

		// Add basicProjectID - topLevelProjectName into hierarchy.
		hierarchy[project.ID] = topLevelProjectName
	}

	return hierarchy
}
