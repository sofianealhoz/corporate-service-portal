package service

import "testing"

// table-driven : aucune base de données, on ne teste que les règles métier
func TestCreateInputValidate(t *testing.T) {
	valid := CreateInput{
		Slug:      "audit-securite",
		Title:     "Audit de sécurité",
		Tier:      "standard",
		DurationH: 14,
	}

	cases := []struct {
		name    string
		input   CreateInput
		wantErr bool
	}{
		{"entrée valide", valid, false},
		// on copie valid puis on casse un seul champ
		{"slug manquant", func() CreateInput { c := valid; c.Slug = ""; return c }(), true},
		{"titre manquant", func() CreateInput { c := valid; c.Title = ""; return c }(), true},
		{"gamme inconnue", func() CreateInput { c := valid; c.Tier = "gold"; return c }(), true},
		{"durée nulle", func() CreateInput { c := valid; c.DurationH = 0; return c }(), true},
		{"durée négative", func() CreateInput { c := valid; c.DurationH = -3; return c }(), true},
		{"prix négatif", func() CreateInput { c := valid; c.PriceCents = -100; return c }(), true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()
			if tc.wantErr && err == nil {
				t.Fatalf("une erreur était attendue")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("aucune erreur attendue, obtenu : %v", err)
			}
		})
	}
}
