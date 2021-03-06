/*
Copyright 2021 The Skaffold Authors

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

package access

import (
	"sync"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/deploy/label"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubectl"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/kubernetes/portforward"
)

type Provider interface {
	GetKubernetesAccessor(*kubernetes.ImageList) Accessor
	GetNoopAccessor() Accessor
}

type fullProvider struct {
	kubernetesAccessor func(*kubernetes.ImageList) Accessor
}

var (
	provider *fullProvider
	once     sync.Once
)

func NewAccessorProvider(config portforward.Config, labelConfig label.Config, cli *kubectl.CLI) Provider {
	once.Do(func() {
		provider = &fullProvider{
			kubernetesAccessor: func(podSelector *kubernetes.ImageList) Accessor {
				if !config.PortForwardOptions().Enabled() {
					return &NoopAccessor{}
				}

				return portforward.NewForwarderManager(cli,
					podSelector,
					labelConfig.RunIDSelector(),
					config.Mode(),
					config.PortForwardOptions(),
					config.PortForwardResources())
			},
		}
	})
	return provider
}

func (p *fullProvider) GetKubernetesAccessor(s *kubernetes.ImageList) Accessor {
	return p.kubernetesAccessor(s)
}

func (p *fullProvider) GetNoopAccessor() Accessor {
	return &NoopAccessor{}
}

// NoopProvider is used in tests
type NoopProvider struct{}

func (p *NoopProvider) GetKubernetesAccessor(_ *kubernetes.ImageList) Accessor {
	return &NoopAccessor{}
}

func (p *NoopProvider) GetNoopAccessor() Accessor {
	return &NoopAccessor{}
}
