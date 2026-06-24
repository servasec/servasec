package features

var F *Registry

type Registry struct {
	enabled map[string]bool
}

func NewRegistry(enabled []string) *Registry {
	m := make(map[string]bool, len(enabled))
	for _, f := range enabled {
		m[f] = true
	}
	return &Registry{enabled: m}
}

func (r *Registry) IsEnabled(name string) bool {
	return r.enabled[name]
}

func (r *Registry) EnabledFeatures() []string {
	out := make([]string, 0, len(r.enabled))
	for f := range r.enabled {
		out = append(out, f)
	}
	return out
}

func Init(licenseKey string) *Registry {
	pro := ParseLicense(licenseKey)
	if len(pro) > 0 {
		F = NewRegistry(append(FreeFeatures(), pro...))
	} else {
		F = NewRegistry(FreeFeatures())
	}
	return F
}
