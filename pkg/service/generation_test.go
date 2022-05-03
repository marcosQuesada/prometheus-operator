package service

import "testing"

func TestAddGenerationVersionOnCache(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	value := int64(10)
	c := NewGenerationCache()
	c.Set(namespace, name, value)
	if expected, got := value, c.Get(namespace, name); expected != got {
		t.Errorf("value do not match, expected %d got %d", expected, got)
	}
}

func TestRemoveGenerationVersionOnCache(t *testing.T) {
	namespace := "default"
	name := "prometheus-server-crd"
	value := int64(10)
	c := NewGenerationCache()
	c.Set(namespace, name, value)
	c.Remove(namespace, name)

	if expected, got := int64(0), c.Get(namespace, name); expected != got {
		t.Errorf("value do not match, expected %d got %d", expected, got)
	}
}
