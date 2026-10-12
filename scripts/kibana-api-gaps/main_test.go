package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSelectResourcesRejectsUnknownNames(t *testing.T) {
	all := []Resource{{Name: "a"}, {Name: "b"}}

	got, err := selectResources(all, []string{"b"})
	require.NoError(t, err)
	assert.Equal(t, []Resource{{Name: "b"}}, got)

	_, err = selectResources(all, []string{"typo"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "typo")

	got, err = selectResources(all, nil)
	require.NoError(t, err)
	assert.Equal(t, all, got)
}

func TestFindPackageBySuffixRequiresExactlyOneMatch(t *testing.T) {
	mod := loadFixture(t, "testdata/fixture")

	path, err := findPackageBySuffix(mod, "/kbapi")
	require.NoError(t, err)
	assert.Equal(t, fixtureKbapi, path)

	_, err = findPackageBySuffix(mod, "/missing")
	require.Error(t, err)
	_, err = findPackageBySuffix(mod, "s")
	require.Error(t, err)
}
