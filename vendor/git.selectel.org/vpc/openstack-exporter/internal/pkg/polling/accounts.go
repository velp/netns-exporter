package polling

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling/cache"
)

// AccountObjects contains OpenStack API objects count per project in
// a single account.
type AccountObjects struct {
	AccountName string
	Region      string
	ProjectID   string
	Count       int
}

// getAccountObjects builds flattened list of account objects from the following
// mapping:
//   project_id_1:
//     region_1: objects_count
//	   region_2: objects_count
//	 project_id_2:
//     region_1: objects_count
//	   region_2: objects_count
func getAccountObjects(projectsObjects map[string]map[string]int) []*AccountObjects {
	// Populate slice of account objects. We don't care if there are identical
	// account names between different projects because we don't need to
	// represent two-level identity structure here.
	projectsHierarchy := cache.GetProjectsHierarchy()
	accountsObjects := []*AccountObjects{}

	for projectID, regions := range projectsObjects {
		// Try to retrieve account name from the project hierarchy and skip
		// projects with unknown account names.
		accountName, ok := projectsHierarchy[projectID]
		if !ok {
			log.Warnf(
				"can't get account name for %s project, won't calculate objects count for it",
				projectID,
			)
			continue
		}

		// Skip project with an unknown account name.
		if accountName == "" {
			log.Warnf(
				"got empty account name for %s project, won't calculate objects count for it",
				projectID,
			)
			continue
		}

		for regionName, regionObjectsCount := range regions {
			// Add new AccountObjects data entity.
			accountsObjects = append(accountsObjects, &AccountObjects{
				AccountName: accountName,
				Region:      regionName,
				ProjectID:   projectID,
				Count:       regionObjectsCount,
			})
		}
	}

	return accountsObjects
}
