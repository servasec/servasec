package features

var F *Registry

// ProImplsLoaded is set to true by the private servasec-pro module's init().
// When false (community build), Init() ignores the license key and returns only free features.
var ProImplsLoaded bool

type Registry struct {
	enabled map[string]bool
	quota   *Quota
}

func NewRegistry(enabled []string) *Registry {
	return NewRegistryWithQuota(enabled, nil)
}

func NewRegistryWithQuota(enabled []string, quota *Quota) *Registry {
	m := make(map[string]bool, len(enabled))
	for _, f := range enabled {
		m[f] = true
	}
	return &Registry{enabled: m, quota: quota}
}

func (r *Registry) IsEnabled(name string) bool {
	return r.enabled[name]
}

// Quota returns the quota limits carried by the license, or nil when
// unlimited (community build or license without quota claims).
func (r *Registry) Quota() *Quota {
	return r.quota
}

func (r *Registry) EnabledFeatures() []string {
	out := make([]string, 0, len(r.enabled))
	for f := range r.enabled {
		out = append(out, f)
	}
	return out
}

func Init(licenseKey string) *Registry {
	if !ProImplsLoaded {
		F = NewRegistry(FreeFeatures())
		return F
	}
	pl := ParseLicenseFull(licenseKey)
	if pl != nil {
		F = NewRegistryWithQuota(append(FreeFeatures(), pl.Features...), pl.Quota)
	} else {
		F = NewRegistry(FreeFeatures())
	}
	return F
}
