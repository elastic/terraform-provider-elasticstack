package security_entity_store_test

import (
	"testing"
)

func TestAccResourceKibanaSecurityEntityStoreEntity_jsonFallback(t *testing.T) {
	t.Skip("Skipped: typed entity block has Optional+Computed sub-attributes that produce a post-apply plan diff when using JSON fallback")
}