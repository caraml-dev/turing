//go:build integration

package service

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/database"
	"github.com/caraml-dev/turing/api/turing/internal/ref"
	"github.com/caraml-dev/turing/api/turing/models"
)

func assertEqualEnsembler(
	t *testing.T,
	expected models.EnsemblerLike,
	actual models.EnsemblerLike,
) {
	require.Equal(t, reflect.TypeOf(expected), reflect.TypeOf(actual))

	assert.Equal(t, expected.GetID(), actual.GetID())
	assert.Equal(t, expected.GetType(), actual.GetType())
	assert.Equal(t, expected.GetName(), actual.GetName())
	assert.Equal(t, expected.GetProjectID(), actual.GetProjectID())
	assert.NotNil(t, actual.GetCreatedAt())
	assert.NotNil(t, actual.GetUpdatedAt())

	switch actual.(type) {
	case *models.PyFuncEnsembler:
		expectedConcrete := expected.(*models.PyFuncEnsembler)
		actualConcrete := actual.(*models.PyFuncEnsembler)
		assert.Equal(t, models.EnsemblerPyFuncType, actualConcrete.GetType())
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
		otherProjectID := models.ID(2)
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

		// add one more ensembler with another ProjectID
		ensemblers = append(ensemblers,
			&models.PyFuncEnsembler{
				GenericEnsembler: &models.GenericEnsembler{
					ProjectID: otherProjectID,
					Name:      fmt.Sprintf("test-ensembler-%d", numEnsemblers),
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
			ProjectID: &otherProjectID,
		})
		assert.NoError(t, err)
		assertEqualEnsembler(t, ensemblers[numEnsemblers], actual)

		// List First Page
		pageSize := 6
		pages := int(math.Ceil(float64(numEnsemblers) / float64(pageSize)))
		fetched, err := svc.List(
			EnsemblersListOptions{
				PaginationOptions: PaginationOptions{
					Page:     ref.Int(1),
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
					Page:     ref.Int(2),
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
					Page:     ref.Int(3),
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

		// Insert non-pyfunc ensembler
		_, err = svc.Save(&models.GenericEnsembler{
			ProjectID: projectID,
			Type:      "unknown",
			Name:      "unknown-ensembler",
		})
		require.NoError(t, err)

		// Fetch all of specific type
		pyfunc := models.EnsemblerPyFuncType
		fetched, err = svc.List(EnsemblersListOptions{
			ProjectID:     &projectID,
			EnsemblerType: &pyfunc,
		})
		require.NoError(t, err)
		assert.Equal(t, numEnsemblers, fetched.Paging.Total)
		results, ok = fetched.Results.([]*models.GenericEnsembler)
		assert.ElementsMatch(t, found, results)
	})
}
