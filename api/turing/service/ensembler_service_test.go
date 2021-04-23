// +build integration

package service

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/gojek/turing/api/turing/it/database"
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func assertEqualEnsembler(
	t *testing.T,
	expected models.EnsemblerLike,
	actual models.EnsemblerLike,
) {
	require.Equal(t, reflect.TypeOf(expected), reflect.TypeOf(actual))

	assertGeneric := func(t *testing.T, expected *models.GenericEnsembler, actual *models.GenericEnsembler) {
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.ProjectID, actual.ProjectID)
		assert.NotNil(t, actual.CreatedAt)
		assert.NotNil(t, actual.UpdatedAt)
	}

	switch actual.(type) {
	case *models.PyFuncEnsembler:
		expectedConcrete := expected.(*models.PyFuncEnsembler)
		actualConcrete := actual.(*models.PyFuncEnsembler)
		assertGeneric(t, expectedConcrete.GenericEnsembler, actualConcrete.GenericEnsembler)
		assert.Equal(t, models.EnsemblerTypePyFunc, actualConcrete.Type)
		assert.Equal(t, expectedConcrete.ExperimentID, actualConcrete.ExperimentID)
		assert.Equal(t, expectedConcrete.RunID, actualConcrete.RunID)
		assert.Equal(t, expectedConcrete.ArtifactURI, actualConcrete.ArtifactURI)
	}

}

func TestEnsemblersServiceIntegration(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		svc := NewEnsemblersService(db)

		numEnsemblers := 10
		projectID := models.ID(1)
		ensemblers := make([]models.EnsemblerLike, numEnsemblers)
		for i := 0; i < numEnsemblers; i++ {
			ensemblers[i] = &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					ProjectID: projectID,
					Name:      fmt.Sprintf("test-ensembler-%d", i),
				},
				ExperimentID: models.ID(10 + i),
				RunID:        fmt.Sprintf("experiment-run-%d", i),
				ArtifactURI:  fmt.Sprintf("gs://bucket-name/ensembler/%d/artifacts", i),
			}
		}

		// Create ensemblers
		for _, ensembler := range ensemblers {
			saved, err := svc.Save(ensembler)
			assert.NoError(t, err)
			assertEqualEnsembler(t, ensembler, saved)
		}

		// Fetch ensembler by its ID
		found := make([]*models.GenericEnsembler, numEnsemblers)
		for i := 0; i < numEnsemblers; i++ {
			actual, err := svc.FindByID(models.ID(i + 1))
			assert.NoError(t, err)
			assertEqualEnsembler(t, ensemblers[i], actual)
			switch actual.(type) {
			case *models.PyFuncEnsembler:
				found[i] = actual.(*models.PyFuncEnsembler).GenericEnsembler
			}
		}

		// List First Page
		pageSize := 6
		pages := int(math.Ceil(float64(numEnsemblers) / float64(pageSize)))
		fetched, err := svc.List(projectID, ListEnsemblersQuery{
			paginationQuery{
				page:     1,
				pageSize: pageSize,
			},
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, fetched.Paging.Page)
		assert.Equal(t, numEnsemblers, fetched.Paging.Total)
		assert.Equal(t, pages, fetched.Paging.Pages)
		results, ok := fetched.Results.([]*models.GenericEnsembler)
		assert.True(t, ok)
		assert.ElementsMatch(t, found[0:pageSize], results)

		// Next Page
		fetched, err = svc.List(projectID, ListEnsemblersQuery{
			paginationQuery{
				page:     2,
				pageSize: pageSize,
			},
		})
		assert.NoError(t, err)
		results, ok = fetched.Results.([]*models.GenericEnsembler)
		assert.Equal(t, numEnsemblers-pageSize, len(results))
		assert.True(t, ok)
		assert.ElementsMatch(t, found[pageSize:], results)

		// Empty results
		fetched, err = svc.List(projectID, ListEnsemblersQuery{
			paginationQuery{
				page:     3,
				pageSize: pageSize,
			},
		})
		assert.NoError(t, err)
		results, ok = fetched.Results.([]*models.GenericEnsembler)
		assert.True(t, ok)
		assert.Empty(t, results)
	})
}
