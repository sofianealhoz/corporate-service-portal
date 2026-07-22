-- Fichiers appliqués dans l'ordre alphabétique, une seule fois chacun.

CREATE TABLE IF NOT EXISTS services (
    id          BIGSERIAL   PRIMARY KEY,

    -- identifiant lisible dans l'URL, plutôt que l'id numérique
    slug        TEXT        NOT NULL UNIQUE,

    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',

    -- contraint en base : dernier rempart si la validation applicative rate
    tier        TEXT        NOT NULL CHECK (tier IN ('standard', 'premium', 'enterprise')),

    duration_h  INTEGER     NOT NULL CHECK (duration_h > 0),

    -- en centimes, jamais en flottant (erreurs d'arrondi sur des montants)
    price_cents INTEGER     NOT NULL DEFAULT 0 CHECK (price_cents >= 0),

    published   BOOLEAN     NOT NULL DEFAULT FALSE,

    -- TIMESTAMPTZ : instant absolu, indépendant du fuseau du serveur
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- couvre la requête la plus fréquente : publiés, les plus récents d'abord
CREATE INDEX IF NOT EXISTS idx_services_published
    ON services (published, created_at DESC);
