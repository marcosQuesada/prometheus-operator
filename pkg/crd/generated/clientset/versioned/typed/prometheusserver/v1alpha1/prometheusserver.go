/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	"time"

	v1alpha1 "github.com/marcosQuesada/prometheus-operator/pkg/crd/apis/prometheusserver/v1alpha1"
	scheme "github.com/marcosQuesada/prometheus-operator/pkg/crd/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PrometheusServersGetter has a method to return a PrometheusServerInterface.
// A group's client should implement this interface.
type PrometheusServersGetter interface {
	PrometheusServers(namespace string) PrometheusServerInterface
}

// PrometheusServerInterface has methods to work with PrometheusServer resources.
type PrometheusServerInterface interface {
	Create(ctx context.Context, prometheusServer *v1alpha1.PrometheusServer, opts v1.CreateOptions) (*v1alpha1.PrometheusServer, error)
	Update(ctx context.Context, prometheusServer *v1alpha1.PrometheusServer, opts v1.UpdateOptions) (*v1alpha1.PrometheusServer, error)
	UpdateStatus(ctx context.Context, prometheusServer *v1alpha1.PrometheusServer, opts v1.UpdateOptions) (*v1alpha1.PrometheusServer, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.PrometheusServer, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.PrometheusServerList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PrometheusServer, err error)
	PrometheusServerExpansion
}

// prometheusServers implements PrometheusServerInterface
type prometheusServers struct {
	client rest.Interface
	ns     string
}

// newPrometheusServers returns a PrometheusServers
func newPrometheusServers(c *K8slabV1alpha1Client, namespace string) *prometheusServers {
	return &prometheusServers{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the prometheusServer, and returns the corresponding prometheusServer object, and an error if there is any.
func (c *prometheusServers) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.PrometheusServer, err error) {
	result = &v1alpha1.PrometheusServer{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("prometheusservers").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PrometheusServers that match those selectors.
func (c *prometheusServers) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.PrometheusServerList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.PrometheusServerList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("prometheusservers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested prometheusServers.
func (c *prometheusServers) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("prometheusservers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a prometheusServer and creates it.  Returns the server's representation of the prometheusServer, and an error, if there is any.
func (c *prometheusServers) Create(ctx context.Context, prometheusServer *v1alpha1.PrometheusServer, opts v1.CreateOptions) (result *v1alpha1.PrometheusServer, err error) {
	result = &v1alpha1.PrometheusServer{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("prometheusservers").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(prometheusServer).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a prometheusServer and updates it. Returns the server's representation of the prometheusServer, and an error, if there is any.
func (c *prometheusServers) Update(ctx context.Context, prometheusServer *v1alpha1.PrometheusServer, opts v1.UpdateOptions) (result *v1alpha1.PrometheusServer, err error) {
	result = &v1alpha1.PrometheusServer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("prometheusservers").
		Name(prometheusServer.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(prometheusServer).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *prometheusServers) UpdateStatus(ctx context.Context, prometheusServer *v1alpha1.PrometheusServer, opts v1.UpdateOptions) (result *v1alpha1.PrometheusServer, err error) {
	result = &v1alpha1.PrometheusServer{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("prometheusservers").
		Name(prometheusServer.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(prometheusServer).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the prometheusServer and deletes it. Returns an error if one occurs.
func (c *prometheusServers) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("prometheusservers").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *prometheusServers) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("prometheusservers").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched prometheusServer.
func (c *prometheusServers) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.PrometheusServer, err error) {
	result = &v1alpha1.PrometheusServer{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("prometheusservers").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
