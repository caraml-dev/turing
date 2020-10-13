package service

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
	"github.com/xanzy/go-gitlab"
	"gotest.tools/assert"
)

func TestGitlabOpsAlertServiceSave(t *testing.T) {
	mockGitlab, requestRecords := newMockGitlabServer()
	defer mockGitlab.Close()

	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	mockSQL.ExpectBegin()
	mockSQL.
		ExpectQuery(`INSERT INTO "alerts"`).
		WithArgs(time.Unix(1593647218, 0), time.Unix(1593647219, 0), "env", "team",
			"service", "throughput", float64(50), float64(25), "5m").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mockSQL.ExpectCommit()
	mockSQL.
		ExpectQuery(`SELECT (.+) FROM "alerts"`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	git, err := gitlab.NewClient("token", gitlab.WithBaseURL(mockGitlab.URL))
	assert.NilError(t, err)

	service := NewGitlabOpsAlertService(mockDb, git, "project", "master", "prefix")
	alert := models.Alert{
		Model: models.Model{
			CreatedAt: time.Unix(1593647218, 0),
			UpdatedAt: time.Unix(1593647219, 0),
		},
		Environment:       "env",
		Team:              "team",
		Service:           "service",
		Metric:            models.MetricThroughput,
		WarningThreshold:  50,
		CriticalThreshold: 25,
		Duration:          "5m",
	}

	_, err = service.Save(alert, "user@gojek.com")
	assert.NilError(t, err)

	err = mockSQL.ExpectationsWereMet()
	assert.NilError(t, err)

	expected := "POST /api/v4/projects/project/repository/files/prefix/env/team/service/throughput.yaml"
	_, ok := requestRecords[expected]
	assert.Check(t, ok == true, "Request: %s is not called. Actual requests: %v", expected, requestRecords)
}

func TestGitlabOpsAlertServiceSaveShouldRevertGitWhenDbFail(t *testing.T) {
	mockGitlab, requestRecords := newMockGitlabServer()
	defer mockGitlab.Close()

	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	mockSQL.ExpectBegin()
	mockSQL.
		ExpectQuery(`INSERT INTO "alerts"`).
		WillReturnError(errors.New("insertion error"))

	git, err := gitlab.NewClient("token", gitlab.WithBaseURL(mockGitlab.URL))
	assert.NilError(t, err)

	service := NewGitlabOpsAlertService(mockDb, git, "project", "master", "prefix")
	alert := models.Alert{
		Environment: "env",
		Team:        "team",
		Service:     "service",
		Metric:      models.MetricThroughput,
		Duration:    "5m",
	}

	_, err = service.Save(alert, "user@gojek.com")
	assert.ErrorContains(t, err, "insertion error")

	expected := "DELETE /api/v4/projects/project/repository/files/prefix/env/team/service/throughput.yaml"
	_, ok := requestRecords[expected]
	assert.Check(t, ok == true, "Request: %s is not called. Actual requests: %v", expected, requestRecords)
}

func TestGitlabOpsAlertServiceList(t *testing.T) {
	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	columns := []string{"environment", "team", "service", "metric", "duration"}
	mockSQL.
		ExpectQuery(`SELECT (.+) FROM "alerts"`).
		WithArgs("service").
		WillReturnRows(sqlmock.NewRows(columns).AddRow("env", "team", "service", "throughput", "5m"))

	service := NewGitlabOpsAlertService(mockDb, nil, "", "", "")

	alerts, err := service.List("service")
	assert.NilError(t, err)
	assert.Equal(t, len(alerts), 1)
	assert.DeepEqual(t, alerts[0],
		&models.Alert{
			Environment: "env",
			Team:        "team",
			Service:     "service",
			Metric:      "throughput",
			Duration:    "5m",
		},
	)

	err = mockSQL.ExpectationsWereMet()
	assert.NilError(t, err)
}

func TestGitlabOpsAlertServiceFindByID(t *testing.T) {
	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	columns := []string{"environment", "team", "service", "metric", "duration"}
	mockSQL.
		ExpectQuery(`SELECT (.+) FROM "alerts"`).
		WithArgs(5).
		WillReturnRows(sqlmock.NewRows(columns).AddRow("env", "team", "service", "throughput", "5m"))

	service := NewGitlabOpsAlertService(mockDb, nil, "", "", "")

	alert, err := service.FindByID(5)
	assert.NilError(t, err)
	assert.DeepEqual(t, alert,
		&models.Alert{
			Environment: "env",
			Team:        "team",
			Service:     "service",
			Metric:      "throughput",
			Duration:    "5m",
		},
	)

	err = mockSQL.ExpectationsWereMet()
	assert.NilError(t, err)
}

func TestGitlabOpsAlertServiceFindByIDShouldReturnErrWhenNotFound(t *testing.T) {
	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	mockSQL.
		ExpectQuery(`SELECT (.+) FROM "alerts"`).
		WithArgs(1).
		WillReturnError(errors.New("select not found"))

	service := NewGitlabOpsAlertService(mockDb, nil, "", "", "")

	_, err := service.FindByID(1)
	assert.ErrorContains(t, err, "select not found")
}

func TestGitlabOpsAlertServiceUpdate(t *testing.T) {
	mockGitlab, requestRecords := newMockGitlabServer()
	defer mockGitlab.Close()

	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	mockSQL.
		ExpectQuery(`SELECT (.+) FROM "alerts"`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mockSQL.ExpectBegin()
	mockSQL.
		ExpectExec(`UPDATE "alerts"`).
		WithArgs(sqlmock.AnyArg(), "env", "team", "service", "throughput", float64(50), float64(25), "5m", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectCommit()

	git, err := gitlab.NewClient("token", gitlab.WithBaseURL(mockGitlab.URL))
	assert.NilError(t, err)

	service := NewGitlabOpsAlertService(mockDb, git, "project", "master", "prefix")
	alert := models.Alert{
		Model:             models.Model{ID: 1},
		Environment:       "env",
		Team:              "team",
		Service:           "service",
		Metric:            models.MetricThroughput,
		WarningThreshold:  50,
		CriticalThreshold: 25,
		Duration:          "5m",
	}

	err = service.Update(alert, "user@gojek.com")
	assert.NilError(t, err)

	err = mockSQL.ExpectationsWereMet()
	assert.NilError(t, err)

	expected := "PUT /api/v4/projects/project/repository/files/prefix/env/team/service/throughput.yaml"
	_, ok := requestRecords[expected]
	assert.Check(t, ok == true, "Request: %s is not called. Actual requests: %v", expected, requestRecords)
}

func TestGitlabOpsAlertServiceUpdateShouldRevertGitWhenDbFail(t *testing.T) {
	mockGitlab, requestRecords := newMockGitlabServer()
	defer mockGitlab.Close()

	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	mockSQL.
		ExpectQuery(`SELECT (.+) FROM "alerts"`).
		WithArgs(1).
		WillReturnRows(
			sqlmock.
				NewRows([]string{"id", "environment", "team", "service", "metric"}).
				AddRow(1, "env", "team", "service", models.MetricLatency95p),
		)
	mockSQL.ExpectBegin()
	mockSQL.
		ExpectExec(`UPDATE "alerts"`).
		WillReturnError(errors.New("update error"))

	git, err := gitlab.NewClient("token", gitlab.WithBaseURL(mockGitlab.URL))
	assert.NilError(t, err)

	service := NewGitlabOpsAlertService(mockDb, git, "project", "master", "prefix")
	alert := models.Alert{
		Model:       models.Model{ID: 1},
		Environment: "env",
		Team:        "team",
		Service:     "service",
		Metric:      models.MetricThroughput,
		Duration:    "5m",
	}

	err = service.Update(alert, "user@gojek.com")
	assert.ErrorContains(t, err, "update error")

	expected := "PUT /api/v4/projects/project/repository/files/prefix/env/team/service/throughput.yaml"
	_, ok := requestRecords[expected]
	assert.Check(t, ok == true, "Request: %s is not called. Actual requests: %v", expected, requestRecords)

	expected = "PUT /api/v4/projects/project/repository/files/prefix/env/team/service/latency95p.yaml"
	_, ok = requestRecords[expected]
	assert.Check(t, ok == true, "Request: %s is not called. Actual requests: %v", expected, requestRecords)
}

func TestGitlabOpsAlertSeviceDelete(t *testing.T) {
	mockGitlab, requestRecords := newMockGitlabServer()
	defer mockGitlab.Close()

	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	mockSQL.
		ExpectQuery(`SELECT (.+) FROM "alerts"`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mockSQL.ExpectBegin()
	mockSQL.
		ExpectExec(`DELETE FROM "alerts"`).
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mockSQL.ExpectCommit()

	git, err := gitlab.NewClient("token", gitlab.WithBaseURL(mockGitlab.URL))
	assert.NilError(t, err)

	service := NewGitlabOpsAlertService(mockDb, git, "project", "master", "prefix")
	alert := models.Alert{
		Model: models.Model{
			ID: 1,
		},
		Environment: "env",
		Team:        "team",
		Service:     "service",
		Metric:      models.MetricLatency95p,
		Duration:    "5m",
	}

	err = service.Delete(alert, "user@gojek.com")
	assert.NilError(t, err)

	err = mockSQL.ExpectationsWereMet()
	assert.NilError(t, err)

	expected := "DELETE /api/v4/projects/project/repository/files/prefix/env/team/service/latency95p.yaml"
	_, ok := requestRecords[expected]
	assert.Check(t, ok == true, "Request: %s is not called. Actual requests: %v", expected, requestRecords)
}

func TestGitlabOpsAlertSeviceDeleteShouldRevertGitWhenDbFail(t *testing.T) {
	mockGitlab, requestRecords := newMockGitlabServer()
	defer mockGitlab.Close()

	mockDb, mockSQL := newMockSQL(t)
	defer mockDb.Close()

	mockSQL.
		ExpectQuery(`SELECT (.+) FROM "alerts"`).
		WithArgs(1).
		WillReturnRows(
			sqlmock.
				NewRows([]string{"id", "environment", "team", "service", "metric"}).
				AddRow(1, "env", "team", "service", models.MetricLatency95p),
		)
	mockSQL.ExpectBegin()
	mockSQL.
		ExpectExec(`DELETE FROM "alerts"`).
		WillReturnError(errors.New("delete error"))

	git, err := gitlab.NewClient("token", gitlab.WithBaseURL(mockGitlab.URL))
	assert.NilError(t, err)

	service := NewGitlabOpsAlertService(mockDb, git, "project", "master", "prefix")
	alert := models.Alert{
		Model:       models.Model{ID: 1},
		Environment: "env",
		Team:        "team",
		Service:     "service",
		Metric:      models.MetricLatency95p,
		Duration:    "5m",
	}

	err = service.Delete(alert, "user@gojek.com")
	assert.ErrorContains(t, err, "delete error")

	expected := "DELETE /api/v4/projects/project/repository/files/prefix/env/team/service/latency95p.yaml"
	_, ok := requestRecords[expected]
	assert.Check(t, ok == true, "Request: %s is not called. Actual requests: %v", expected, requestRecords)

	expected = "POST /api/v4/projects/project/repository/files/prefix/env/team/service/latency95p.yaml"
	_, ok = requestRecords[expected]
	assert.Check(t, ok == true, "Request: %s is not called. Actual requests: %v", expected, requestRecords)
}

func newMockGitlabServer() (mockGitlab *httptest.Server, requestRecords map[string]int) {
	records := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		records[record]++

		if strings.HasPrefix(r.URL.Path, "/api/v4/projects/project/repository/files/") && r.Method != http.MethodGet {
			fmt.Fprintf(w, `{"file_path":"file_path.yaml","branch":"master"}`)
		} else {
			fmt.Fprintf(w, "")
		}
	}))
	return server, records
}

func newMockSQL(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	mockdb, mocksql, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open stub database connection: %s", err)
	}
	gormDb, err := gorm.Open("postgres", mockdb)
	if err != nil {
		t.Fatalf("failed to open go-orm database connection: %s", err)
	}
	return gormDb, mocksql
}
