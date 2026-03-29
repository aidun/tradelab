package main

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestApplyMigrationCommitsSchemaAndVersionInOneTransaction(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE demo").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO schema_migrations").WithArgs("000001_init").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err = applyMigration(context.Background(), db, migrationFile{
		version:  "000001_init",
		contents: "CREATE TABLE demo (id INT);",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplyMigrationRollsBackWhenRecordingVersionFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("CREATE TABLE demo").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO schema_migrations").WithArgs("000001_init").WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	err = applyMigration(context.Background(), db, migrationFile{
		version:  "000001_init",
		contents: "CREATE TABLE demo (id INT);",
	})
	if err == nil {
		t.Fatal("expected an error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestApplyPendingMigrationsSkipsAlreadyAppliedVersions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}
	defer db.Close()

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)
	mock.ExpectQuery("SELECT EXISTS").WithArgs("000001_init").WillReturnRows(rows)

	err = applyPendingMigrations(context.Background(), db, []migrationFile{
		{version: "000001_init", contents: "CREATE TABLE demo (id INT);"},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
