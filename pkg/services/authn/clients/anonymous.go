package clients

import (
	"context"
	"strings"

	"github.com/grafana/grafana/pkg/infra/kvstore"
	"github.com/grafana/grafana/pkg/infra/log"
	"github.com/grafana/grafana/pkg/services/authn"
	"github.com/grafana/grafana/pkg/services/org"
	"github.com/grafana/grafana/pkg/setting"
)

var _ authn.ContextAwareClient = new(Anonymous)

func ProvideAnonymous(cfg *setting.Cfg, orgService org.Service, _ kvstore.KVStore) *Anonymous {
	return &Anonymous{
		cfg:        cfg,
		log:        log.New("authn.anonymous"),
		orgService: orgService,
	}
}

type Anonymous struct {
	cfg        *setting.Cfg
	log        log.Logger
	orgService org.Service
}

func (a *Anonymous) Name() string {
	return authn.ClientAnonymous
}

func (a *Anonymous) Authenticate(ctx context.Context, r *authn.Request) (*authn.Identity, error) {
	o, err := a.orgService.GetByName(ctx, &org.GetOrgByNameQuery{Name: a.cfg.AnonymousOrgName})
	if err != nil {
		a.log.FromContext(ctx).Error("failed to find organization", "name", a.cfg.AnonymousOrgName, "error", err)
		return nil, err
	}

	return &authn.Identity{
		IsAnonymous:  true,
		OrgID:        o.ID,
		OrgName:      o.Name,
		OrgRoles:     map[int64]org.RoleType{o.ID: org.RoleType(a.cfg.AnonymousOrgRole)},
		ClientParams: authn.ClientParams{},
	}, nil
}

func (a *Anonymous) Test(ctx context.Context, r *authn.Request) bool {
	// If anonymous client is register it can always be used for authentication
	return true
}

func (a *Anonymous) Priority() uint {
	return 100
}

func (a *Anonymous) UsageStatFn(ctx context.Context) (map[string]interface{}, error) {
	m := map[string]interface{}{}

	// Add stats about anonymous auth
	m["stats.anonymous.customized_role.count"] = 0
	if !strings.EqualFold(a.cfg.AnonymousOrgRole, "Viewer") {
		m["stats.anonymous.customized_role.count"] = 1
	}

	return m, nil
}
