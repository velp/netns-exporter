package projects

import (
	"fmt"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/utils"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/domains"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
	"github.com/gophercloud/gophercloud/pagination"
)

// GetProjects retrieves projects from the OpenStack Identity V3 API.
func GetProjects(opts *common.GetOpts) ([]projects.Project, error) {
	isDomain := false
	allPages, err := utils.RetryForPager(opts.Attempts, opts.Interval, func() (pagination.Page, error) {
		return projects.List(opts.Client, projects.ListOpts{
			IsDomain: &isDomain,
		}).AllPages()
	})
	if err != nil {
		return nil, err
	}
	allProjects, err := projects.ExtractProjects(allPages)
	if err != nil {
		return nil, err
	}

	return allProjects, nil
}

// GetTopLevelProjects retrieves top-level projects from the OpenStack Identity V3 API.
func GetTopLevelProjects(opts *common.GetOpts) ([]domains.Domain, error) {
	allPages, err := utils.RetryForPager(opts.Attempts, opts.Interval, func() (pagination.Page, error) {
		return domains.List(opts.Client, domains.ListOpts{}).AllPages()
	})
	if err != nil {
		return nil, err
	}
	allDomains, err := domains.ExtractDomains(allPages)
	if err != nil {
		return nil, err
	}

	return allDomains, nil
}

// FlattenTopLevelProjects returns the following map from the provided top-level
// projects:
//
// map[string]string {
//   id: name,
// }
func FlattenTopLevelProjects(topLevelProjects []domains.Domain) map[string]string {
	flattenProjects := make(map[string]string, len(topLevelProjects))

	for _, project := range topLevelProjects {
		flattenProjects[project.ID] = project.Name
	}

	return flattenProjects
}

// GetOctaviaProject retrieves Octavia service project from the OpenStack Identity V3 API.
func GetOctaviaProject(opts *common.GetOpts) (*projects.Project, error) {
	allPages, err := utils.RetryForPager(opts.Attempts, opts.Interval, func() (pagination.Page, error) {
		return projects.List(opts.Client, projects.ListOpts{
			DomainID: config.Config.OpenStack.Octavia.DomainName,
			Name:     config.Config.OpenStack.Octavia.ProjectName,
		}).AllPages()
	})
	if err != nil {
		return nil, err
	}
	allProjects, err := projects.ExtractProjects(allPages)
	if err != nil {
		return nil, err
	}
	if len(allProjects) != 1 {
		return nil, fmt.Errorf("want one Octavia project, got %d", len(allProjects))
	}
	return &allProjects[0], nil
}
