package service

import "testing"

// Test « table-driven » : l'idiome standard en Go.
// On déclare une table de cas, puis on boucle dessus.
// Équivalent de @pytest.mark.parametrize en Python.
//
// Note : aucune base de données ici. C'est tout l'intérêt d'avoir séparé les
// règles métier (model) de l'accès aux données (repo) — ce test tourne en
// quelques millisecondes, sans PostgreSQL.
func TestCreateInputValidate(t *testing.T) {
	// valid = un jeu de données correct, qu'on dégrade dans chaque cas.
	valid := CreateInput{
		Slug:      "audit-securite",
		Title:     "Audit de sécurité",
		Tier:      "standard",
		DurationH: 14,
	}

	cases := []struct {
		name    string      // nom du sous-test, affiché en cas d'échec
		input   CreateInput // la donnée testée
		wantErr bool        // attend-on une erreur ?
	}{
		{
			name:    "entrée valide",
			input:   valid,
			wantErr: false,
		},
		{
			name: "slug manquant",
			// On copie `valid` puis on casse UN champ : le test dit alors
			// exactement quelle règle est en cause.
			input:   func() CreateInput { c := valid; c.Slug = ""; return c }(),
			wantErr: true,
		},
		{
			name:    "titre manquant",
			input:   func() CreateInput { c := valid; c.Title = ""; return c }(),
			wantErr: true,
		},
		{
			name:    "gamme inconnue",
			input:   func() CreateInput { c := valid; c.Tier = "gold"; return c }(),
			wantErr: true,
		},
		{
			name:    "durée nulle",
			input:   func() CreateInput { c := valid; c.DurationH = 0; return c }(),
			wantErr: true,
		},
		{
			name:    "durée négative",
			input:   func() CreateInput { c := valid; c.DurationH = -3; return c }(),
			wantErr: true,
		},
		{
			name:    "prix négatif",
			input:   func() CreateInput { c := valid; c.PriceCents = -100; return c }(),
			wantErr: true,
		},
	}

	for _, tc := range cases {
		// t.Run crée un SOUS-TEST nommé : en cas d'échec, la sortie indique
		// précisément quel cas a cassé, et on peut n'en relancer qu'un seul
		// (go test -run 'TestCreateInputValidate/gamme_inconnue').
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.Validate()

			if tc.wantErr && err == nil {
				t.Fatalf("une erreur était attendue, mais Validate() a réussi")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("aucune erreur attendue, obtenu : %v", err)
			}
		})
	}
}
