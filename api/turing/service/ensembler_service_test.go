// +build integration

package service

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/gojek/turing/api/turing/internal/testutils"
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

	assert.Equal(t, expected.Type(), actual.Type())
	assert.Equal(t, expected.Name(), actual.Name())
	assert.Equal(t, expected.ProjectID(), actual.ProjectID())

	assertGeneric := func(t *testing.T, expected *models.GenericEnsembler, actual *models.GenericEnsembler) {
		assert.Equal(t, expected.ID, actual.ID)
		assert.NotNil(t, actual.CreatedAt)
		assert.NotNil(t, actual.UpdatedAt)
	}

	switch actual.(type) {
	case *models.PyFuncEnsembler:
		expectedConcrete := expected.(*models.PyFuncEnsembler)
		actualConcrete := actual.(*models.PyFuncEnsembler)
		assertGeneric(t, expectedConcrete.GenericEnsembler, actualConcrete.GenericEnsembler)
		assert.Equal(t, models.EnsemblerTypePyFunc, actualConcrete.Type())
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
		projectID_2 := models.ID(2)
		ensemblers := make([]models.EnsemblerLike, numEnsemblers)
		for i := 0; i < numEnsemblers; i++ {
			ensemblers[i] = &models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					TProjectID: projectID,
					TName:      fmt.Sprintf("test-ensembler-%d", i),
				},
				ExperimentID: models.ID(10 + i),
				RunID:        fmt.Sprintf("experiment-run-%d", i),
				ArtifactURI:  fmt.Sprintf("gs://bucket-name/ensembler/%d/artifacts", i),
			}
		}

		// add one more ensembler with another ProjectID
		ensemblers = append(ensemblers,
			&models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					TProjectID: projectID_2,
					TName:      fmt.Sprintf("test-ensembler-%d", numEnsemblers),
				},
				ExperimentID: models.ID(10 + numEnsemblers),
				RunID:        fmt.Sprintf("experiment-run-%d", numEnsemblers),
				ArtifactURI:  fmt.Sprintf("gs://bucket-name/ensembler/%d/artifacts", numEnsemblers),
			})

		// Create ensemblers
		for _, ensembler := range ensemblers {
			saved, err := svc.Save(ensembler)
			assert.NoError(t, err)
			assertEqualEnsembler(t, ensembler, saved)
		}

		// Fetch ensembler by its ID
		found := make([]*models.GenericEnsembler, numEnsemblers)
		for i := 0; i < numEnsemblers; i++ {
			actual, err := svc.FindByID(models.ID(i+1), EnsemblersFindByIDOptions{})
			assert.NoError(t, err)
			assertEqualEnsembler(t, ensemblers[i], actual)
			switch actual.(type) {
			case *models.PyFuncEnsembler:
				found[i] = actual.(*models.PyFuncEnsembler).GenericEnsembler
			}
		}

		// Find by ID and ProjectID
		actual, err := svc.FindByID(models.ID(numEnsemblers+1), EnsemblersFindByIDOptions{
			ProjectID: &projectID,
		})
		assert.EqualError(t, err, "record not found")
		assert.Nil(t, actual)

		// Find by ID and ProjectID
		actual, err = svc.FindByID(models.ID(numEnsemblers+1), EnsemblersFindByIDOptions{
			ProjectID: &projectID_2,
		})
		assert.NoError(t, err)
		assertEqualEnsembler(t, ensemblers[numEnsemblers], actual)

		// List First Page
		pageSize := 6
		pages := int(math.Ceil(float64(numEnsemblers) / float64(pageSize)))
		fetched, err := svc.List(
			EnsemblersListOptions{
				PaginationOptions: PaginationOptions{
					Page:     testutils.NullableInt(1),
					PageSize: &pageSize,
				},
				ProjectID: &projectID,
			})
		require.NoError(t, err)
		assert.Equal(t, 1, fetched.Paging.Page)
		assert.Equal(t, numEnsemblers, fetched.Paging.Total)
		assert.Equal(t, pages, fetched.Paging.Pages)
		results, ok := fetched.Results.([]*models.GenericEnsembler)
		require.True(t, ok)
		assert.ElementsMatch(t, found[0:pageSize], results)

		// Next Page
		fetched, err = svc.List(
			EnsemblersListOptions{
				PaginationOptions: PaginationOptions{
					Page:     testutils.NullableInt(2),
					PageSize: &pageSize,
				},
				ProjectID: &projectID,
			})
		require.NoError(t, err)
		results, ok = fetched.Results.([]*models.GenericEnsembler)
		require.True(t, ok)
		assert.Equal(t, numEnsemblers-pageSize, len(results))
		assert.ElementsMatch(t, found[pageSize:], results)

		// Empty results
		fetched, err = svc.List(
			EnsemblersListOptions{
				PaginationOptions: PaginationOptions{
					Page:     testutils.NullableInt(3),
					PageSize: &pageSize,
				},
				ProjectID: &projectID,
			})
		require.NoError(t, err)
		results, ok = fetched.Results.([]*models.GenericEnsembler)
		require.True(t, ok)
		assert.Empty(t, results)

		// Fetch all
		fetched, err = svc.List(EnsemblersListOptions{ProjectID: &projectID})
		require.NoError(t, err)
		assert.Equal(t, 1, fetched.Paging.Page)
		assert.Equal(t, numEnsemblers, fetched.Paging.Total)
		assert.Equal(t, 1, fetched.Paging.Pages)
		results, ok = fetched.Results.([]*models.GenericEnsembler)
		require.True(t, ok)
		assert.Equal(t, numEnsemblers, len(results))
		assert.ElementsMatch(t, found, results)
	})
}
