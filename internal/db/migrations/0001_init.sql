-- Migration 0001 : table des services (la ressource principale du catalogue).
--
-- Convention : les fichiers sont numérotés et appliqués dans l'ordre alphabétique,
-- une seule fois chacun (voir internal/db/migrate.go).

CREATE TABLE IF NOT EXISTS services (
    id          BIGSERIAL   PRIMARY KEY,

    -- 'slug' = identifiant lisible utilisé dans les URLs (/services/audit-securite)
    -- plutôt que l'id numérique : plus propre, et ça n'expose pas le nombre
    -- d'enregistrements. UNIQUE crée automatiquement un index -> la recherche
    -- par slug est rapide.
    slug        TEXT        NOT NULL UNIQUE,

    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',

    -- gamme contrainte côté base : la donnée invalide est refusée même si un bug
    -- passe la validation applicative. La base est le dernier rempart.
    tier        TEXT        NOT NULL CHECK (tier IN ('standard', 'premium', 'enterprise')),

    duration_h  INTEGER     NOT NULL CHECK (duration_h > 0),

    -- Prix stocké en CENTIMES (entier), jamais en flottant : 0.1 + 0.2 != 0.3 en
    -- virgule flottante, ce qui provoque des erreurs d'arrondi sur de l'argent.
    price_cents INTEGER     NOT NULL DEFAULT 0 CHECK (price_cents >= 0),

    -- brouillon vs visible sur le site public
    published   BOOLEAN     NOT NULL DEFAULT FALSE,

    -- TIMESTAMPTZ (avec fuseau) et pas TIMESTAMP : on stocke un instant absolu,
    -- indépendant du fuseau du serveur.
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Index sur la requête la plus fréquente du site public :
-- "les services publiés, les plus récents d'abord".
CREATE INDEX IF NOT EXISTS idx_services_published
    ON services (published, created_at DESC);
