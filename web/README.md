# Front

Interface React du portail de services. Consomme l'API Go du dépôt.

Stack : React, TypeScript, Vite, Chakra UI.

## Lancer en dev

L'API doit tourner d'abord (voir le README racine), sur `:8080`.

```
npm install
npm run dev
```

Le front sert sur `http://localhost:5173`. Vite proxifie `/api` vers
`http://localhost:8080`, donc pas de CORS en dev : le navigateur voit une seule
origine, comme en prod où le back servira le front.

## Build

```
npm run build
```

Sort dans `dist/`.

## Organisation

- `src/types.ts` : types miroir du modèle Go, à garder synchronisés avec l'API.
- `src/api.ts` : seul endroit qui parle à l'API (les écrans ignorent les URLs).
- `src/Catalogue.tsx` : listing des services (`GET /api/services`).
