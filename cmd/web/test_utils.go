package main

import (
	"fireynis/velocity_checker/pkg/models/mock"
	"fireynis/velocity_checker/pkg/validators"
	"testing"
)

func newTestApplication(t *testing.T) *application {
	return &application{
		loads:         &mock.Load{},
		loadValidator: &validators.LoadValidator{},
	}
}
