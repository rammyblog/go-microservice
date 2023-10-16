package main

import (
	"os"
	"testing"

	"github.com/rammyblog/authentication-service/api/data"
)

var testApp Config

func TestMain(m *testing.M) {
	repo := data.NewPostgresTestRepository()
	testApp.Repo = repo
	os.Exit(m.Run())
}
