package crd

import (
	"context"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"testing"
)

func TestEnsureCRDRegisteredOnNonRegisteredCrd(t *testing.T) {
	ini := &fakeInitializer{}
	b := NewBuilder(ini)
	if err := b.EnsureCRDRegistration(context.Background()); err != nil {
		t.Errorf("unable to ensure crd registered, error %v", err)
	}

	if expected, got := 1, ini.creation; expected != got {
		t.Errorf("total calls do not match, expected %d got %d", expected, got)
	}
}

func TestEnsureCRDRegisteredOnAlreadyRegisteredCrd(t *testing.T) {
	ini := &fakeInitializer{result: true}
	b := NewBuilder(ini)
	if err := b.EnsureCRDRegistration(context.Background()); err != nil {
		t.Errorf("unable to ensure crd registered, error %v", err)
	}

	if expected, got := 0, ini.creation; expected != got {
		t.Errorf("total calls do not match, expected %d got %d", expected, got)
	}
}

type fakeInitializer struct {
	result   bool
	error    error
	creation int
}

func (f *fakeInitializer) Create(ctx context.Context, cr *v1.CustomResourceDefinition) error {
	f.creation++
	return f.error
}

func (f *fakeInitializer) IsAccepted(ctx context.Context, resourceName string) (bool, error) {
	return f.result, f.error
}
