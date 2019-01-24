// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package daemon

import (
	"context"

	"github.com/jancajthaml-openbank/bondster-bco-unit/config"
	"github.com/jancajthaml-openbank/bondster-bco-unit/http"

	system "github.com/jancajthaml-openbank/actor-system"
	localfs "github.com/jancajthaml-openbank/local-fs"
)

// ActorSystem represents actor system subroutine
type ActorSystem struct {
	system.Support
	Tenant     string
	Storage    *localfs.Storage
	Metrics    *Metrics
	ClientHTTP http.Client
}

// NewActorSystem returns actor system fascade
func NewActorSystem(ctx context.Context, cfg config.Configuration, metrics *Metrics, storage *localfs.Storage) ActorSystem {
	return ActorSystem{
		Support:    system.NewSupport(ctx, "BondsterUnit/"+cfg.Tenant, cfg.LakeHostname),
		Storage:    storage,
		Tenant:     cfg.Tenant,
		Metrics:    metrics,
		ClientHTTP: http.NewClient(),
	}
}
