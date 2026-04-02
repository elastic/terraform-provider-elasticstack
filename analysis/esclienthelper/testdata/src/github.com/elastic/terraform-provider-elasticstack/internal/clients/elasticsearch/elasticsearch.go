package elasticsearch

import "github.com/elastic/terraform-provider-elasticstack/internal/clients"

func Do(_ *clients.APIClient) error { return nil }

func DoSecond(_ string, _ *clients.APIClient) error { return nil }

func DoFirst(_ *clients.APIClient, _ string) error { return nil }
