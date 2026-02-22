package apikey

import "github.com/elastic/terraform-provider-elasticstack/internal/models"

// populateRoleDescriptorsDefaults ensures that all role descriptors have proper defaults
func populateRoleDescriptorsDefaults(model map[string]models.APIKeyRoleDescriptor) map[string]models.APIKeyRoleDescriptor {
	for role, descriptor := range model {
		resultDescriptor := descriptor

		// Ensure AllowRestrictedIndices is set to false for all indices that don't have it set
		for i, index := range resultDescriptor.Indices {
			if index.AllowRestrictedIndices == nil {
				resultDescriptor.Indices[i].AllowRestrictedIndices = new(bool)
				*resultDescriptor.Indices[i].AllowRestrictedIndices = false
			}
		}

		model[role] = resultDescriptor
	}

	return model
}
