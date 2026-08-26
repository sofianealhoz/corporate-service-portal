# Front

Interface React du portail de services. Consomme l'API Go du dépôt.

Stack : React, TypeScript, Vite, Chakra UI, React Router.

## Lancer en dev

L'API doit tourner d'abord (voir le README racine), sur `:8080`.

```
npm install
npm run dev
```

Le front sert sur `http://localhost:5173`. Vite proxifie `/api` vers
`http://localhost:8080`, donc pas de CORS en dev : le navigateur voit une seule
origine, comme en prod où le back servira le front. Si l'API tourne ailleurs,
`API_PROXY_TARGET` change la cible du proxy.

## Build

```
npm run build
```

Sort dans `dist/`.

## Organisation

Les composants de `src/components/` reçoivent en props ce qu'ils affichent et
n'appellent jamais l'API. C'est ce qui les rend réutilisables : la même carte
sert le catalogue aujourd'hui et une page d'administration demain, sans
modification. Les écrans, eux, chargent les données et les leur passent.

- `src/api.ts` : seul endroit qui parle à l'API (les écrans ignorent les URLs).
  `listServices` accepte `published`, `limit` et `offset` ; `ApiError` porte le
  code HTTP jusqu'aux écrans, qui distinguent un 404 d'une panne.
- `src/types.ts` : types miroir du modèle Go, à garder synchronisés avec l'API.
- `src/format.ts` : formatage des montants, seul endroit qui convertit les
  centimes en euros.
- `src/components/` : `ServiceCard`, `TierBadge`, `Price`, `Pagination` et les
  états `LoadingState`, `EmptyState`, `ErrorState`.
- `src/Catalogue.tsx` : listing (`GET /api/services`), pagination et bascule
  d'affichage des brouillons.
- `src/ServiceDetail.tsx` : fiche sur `/services/{slug}`, avec le cas 404.
