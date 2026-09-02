package features

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestAgentBoundProLicense reproduit le payload exact que l'agent de contrôle
// (servasec-brand/control/agent/license.mjs) émet pour un tenant pro lié au
// slug : claims {features:[pro...], quota:{groups:10,applications:25,
// versions:25,users:0}, iss:servasec, sub:<slug>, exp}. La clé de signature
// étant la même (paire dev), le backend community doit l'accepter quand
// SSC_SITE_NAME matche le sub, et le rejeter sinon.
func TestAgentBoundProLicense(t *testing.T) {
	token := agentProLicense("acme-test")

	t.Setenv("SSC_SITE_NAME", "acme-test")
	pl := ParseLicenseFull(token)
	if pl == nil {
		t.Fatal("license liée avec SSC_SITE_NAME matchant le sub attendue valide")
	}
	if !contains(pl.Features, "sso") {
		t.Errorf("features attendues SIsSO; got %v", pl.Features)
	}
	if pl.Quota == nil || pl.Quota.Groups != 10 || pl.Quota.Applications != 25 ||
		pl.Quota.Versions != 25 || pl.Quota.Users != 0 {
		t.Errorf("quota pro attendu {10,25,25,0}; got %+v", pl.Quota)
	}

	// Slug différent => rejet.
	t.Setenv("SSC_SITE_NAME", "other")
	if pl = ParseLicenseFull(token); pl != nil {
		t.Error("license liée à un autre tenant attendue rejetée")
	}

	// Aucun slug d'instance => rejet.
	os.Unsetenv("SSC_SITE_NAME")
	if pl = ParseLicenseFull(token); pl != nil {
		t.Error("license liée attendue rejetée quand SSC_SITE_NAME est vide")
	}
}

func contains(xs []string, v string) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// agentProLicense construit un JWT ES256 identique au format de l'agent
// (même structure de claims), signé avec la clé dev partagée.
func agentProLicense(subject string) string {
	quota := &Quota{Groups: 10, Applications: 25, Versions: 25, Users: 0}
	claims := licenseClaims{
		Features: []string{"audit_log", "risk_scoring", "advanced_reporting", "mcp_server", "sso"},
		Quota:    quota,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "servasec",
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	signed, err := token.SignedString(parseTestPrivateKey())
	if err != nil {
		panic("failed to sign agent-shaped license: " + err.Error())
	}
	return signed
}