-- La pagination par curseur exige un ordre total : deux services créés dans la
-- même microseconde s'échangeraient sinon entre deux pages, l'un apparaissant
-- deux fois et l'autre jamais. L'id départage, et l'index suit exactement
-- l'ordre demandé par la requête pour qu'elle n'ait rien à trier.

DROP INDEX IF EXISTS idx_services_published;

CREATE INDEX IF NOT EXISTS idx_services_pagination
    ON services (published, created_at DESC, id DESC);
