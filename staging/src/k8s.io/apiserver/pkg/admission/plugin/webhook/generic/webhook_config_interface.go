/*
Copyright 2017 The Kubernetes Authors.

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

package generic

import (
	"k8s.io/api/admissionregistration/v1beta1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/util/webhook"
)

var _ webhook.ClientConfigInterface = &WebhookClientConfigWrapper{}

// WebhookClientConfigWrapper wraps a v1beta1.WebhookClientConfig to support webhook.ClientConfigInterface interface
type WebhookClientConfigWrapper struct {
	v1beta1.WebhookClientConfig
}

// GetURL returns the URL field of embedded WebhookClientConfig
func (c *WebhookClientConfigWrapper) GetURL() *string {
	return c.URL
}

// GetCABundle returns the CABundle field of embedded WebhookClientConfig
func (c *WebhookClientConfigWrapper) GetCABundle() []byte {
	return c.CABundle
}

// GetServiceName returns the Service.Name field of embedded WebhookClientConfig if Service exists, nil otherwise
func (c *WebhookClientConfigWrapper) GetServiceName() *string {
	if c.Service != nil {
		return &c.Service.Name
	}
	return nil
}

// GetServiceNamespace returns the Service.Namespace field of embedded WebhookClientConfig if Service exists, nil otherwise
func (c *WebhookClientConfigWrapper) GetServiceNamespace() *string {
	if c.Service != nil {
		return &c.Service.Namespace
	}
	return nil
}

// GetServicePath returns the Service.Path field of embedded WebhookClientConfig if Service exists, nil otherwise
func (c *WebhookClientConfigWrapper) GetServicePath() *string {
	if c.Service != nil {
		return c.Service.Path
	}
	return nil
}

// GetCacheKey returns a hash key for embedded WebhookClientConfig
func (c *WebhookClientConfigWrapper) GetCacheKey() (string, error) {
	cacheKey, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(cacheKey), nil
}
