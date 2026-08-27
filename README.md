# Corporate Service Portal

Catalogue de services en ligne : vitrine publique, fiche détaillée par service et
espace d'administration. Back-end en Go avec PostgreSQL, front en React.

Reprise et réécriture d'une maquette faite pendant mon stage de développeur web
chez CICAPE. Le dépôt ne couvre que certaines briques du site.

## Stack

Go, chi, pgx, PostgreSQL, Redis, Docker. Front React / TypeScript / Vite / Chakra UI.

## Lancer le projet

Une seule commande, qui démarre la base, le cache, l'API et le front :

```bash
make dev
```

`make help` liste le reste. Le détail, si on préfère lancer chaque brique
soi-même :

L'API :

```bash
docker compose up -d       # PostgreSQL et Redis
go run ./cmd/api           # API sur :8080
curl localhost:8080/health
```

Le front, dans un second terminal :

```bash
cd web
npm install
npm run dev                # front sur :5173
```

Vite proxifie `/api` vers `http://localhost:8080` : le navigateur ne voit qu'une
seule origine, donc pas de CORS en développement. Détails dans
[`web/README.md`](web/README.md).

Variables d'environnement, toutes optionnelles en développement :

| Variable | Défaut | Rôle |
|---|---|---|
| `PORT` | `8080` | port d'écoute |
| `DATABASE_URL` | connexion locale | PostgreSQL |
| `REDIS_URL` | `redis://localhost:6379` | cache |

## Endpoints

| Méthode | Chemin | Description |
|---|---|---|
| `GET` | `/health` | état du service |
| `GET` | `/api/services` | liste, paramètres `published`, `limit`, `offset` |
| `GET` | `/api/services/{slug}` | détail d'un service |
| `POST` | `/api/services` | création |

Un service non publié n'est servi ni par le listing ni par le détail : son slug
répond 404, et non 403, pour ne pas révéler qu'il existe.

Le listing est mis en cache dans Redis pendant 30 secondes et vidé à chaque
création. Si Redis est absent ou en panne, la réponse vient directement de
PostgreSQL.

## Structure

```
cmd/api/           point d'entrée, assemble config, base et routeur
internal/
  config/          lecture des réglages depuis l'environnement
  api/             routeur, handlers, écriture des réponses
  service/         domaine : structure, validation, accès aux données
  db/              pool de connexions et migrations
  cache/           cache Redis, optionnel par construction
  testdb/          base jetable pour les tests d'intégration
web/               front React / TypeScript / Vite, consomme l'API
  src/components/  composants d'affichage, sans accès à l'API
  src/api.ts       seul endroit du front qui connaît les URLs
Makefile           commandes communes aux deux applications
.github/workflows/ CI, un job par application
```

Le paquet `service` contient deux fichiers aux rôles distincts. `model.go` définit
ce qu'est un service et ses règles de validation, sans aucun SQL. `repo.go` est le
seul endroit du projet qui écrit du SQL. Cette séparation permet de tester les
règles métier sans base de données.

Le front suit la même règle. `src/api.ts` est le seul fichier qui connaisse les
URLs de l'API ; les composants de `src/components/` reçoivent en props ce qu'ils
affichent et ne déclenchent aucune requête, ce qui les rend réutilisables et
testables isolément. Deux écrans les assemblent : `Catalogue.tsx`, avec
pagination et bascule des brouillons, et `ServiceDetail.tsx`, sur
`/services/{slug}`.

## Monorepo

Le front et l'API vivent dans le même dépôt. Ce qui en fait un monorepo n'est
pas la cohabitation de deux dossiers, c'est l'outillage qui les tient ensemble :

- **`Makefile` à la racine.** `make dev`, `make build`, `make test`, `make lint`,
  `make clean` marchent quelle que soit la technologie derrière. Personne n'a à
  apprendre deux chaînes d'outils pour contribuer à une moitié du dépôt.
- **Une CI unique**, `.github/workflows/ci.yml`, avec un job par application :
  un job Go qui compile, passe `go vet` et lance les tests sur de vrais
  PostgreSQL et Redis, un job front qui installe, lint et construit.
- **Une convention partagée**, `.editorconfig` à la racine, couvrant les deux.

Pourquoi les réunir : un changement de contrat d'API se voit des deux côtés dans
le même commit. `internal/service/model.go` et `web/src/types.ts` décrivent la
même ressource ; séparés en deux dépôts, ils divergent en silence et la panne
n'apparaît qu'à l'exécution. Ici la revue est unique et les versions ne peuvent
pas se désynchroniser.

Le prix à payer est une CI qui grossit : sans précaution, changer une ligne de
CSS relancerait une base PostgreSQL pour rien. D'où les filtres de chemins, qui
ne déclenchent le job Go que si `cmd/`, `internal/`, les fichiers de modules ou
l'outillage partagé ont bougé, et le job front que pour `web/` ou ce même
outillage. Un job `ci` final agrège les résultats, pour qu'une protection de
branche ne bloque pas sur un job volontairement ignoré.

Ce qui n'a pas été fait : déplacer le Go dans `apps/api/` pour « faire plus
monorepo ». Cela casserait les chemins d'import et les commandes, sans rien
apporter que l'outillage ci-dessus ne donne déjà.

## Tests

```bash
go test ./...
```

Tests unitaires, sans dépendance :

- `internal/service/model_test.go` couvre la validation de `CreateInput` en
  table-driven, sept cas. Aucune base n'est nécessaire, c'est précisément ce que
  permet la séparation décrite au-dessus.
- `internal/cache/cache_test.go` vérifie qu'un cache désactivé, mal configuré ou
  injoignable ne fait jamais échouer un appel.

Tests d'intégration, sur les vrais services du `docker-compose` :

```bash
docker compose up -d
go test ./...
```

- `internal/api/list_test.go` monte le routeur avec `httptest` et parcourt le
  catalogue page par page : chaque service inséré apparaît une fois et une
  seule. Il vérifie aussi qu'une création relue par son slug rend exactement la
  ressource créée.
- `internal/api/services_test.go` vérifie qu'un brouillon répond 404 sur le
  détail.
- `internal/cache/cache_test.go` fait un aller-retour écriture, lecture,
  invalidation sur Redis.

Sans PostgreSQL ni Redis joignables, ces tests sont ignorés, jamais en échec :
`go test ./...` reste vert sur un clone sans Docker.

`internal/testdb` ouvre toujours la base de l'URL suffixée par `_test`, qu'il
crée au besoin, et vide la table `services` avant chaque test. Les données de
développement ne sont donc jamais touchées, même si `TEST_DATABASE_URL` pointe
la base de dev.

Côté front, il n'y a pas encore de tests unitaires. `make test` y lance ce qui
existe : la vérification de types TypeScript et le lint. Les composants sont
écrits pour être testables, ils ne font aucun appel réseau et ne dépendent que
de leurs props, mais le harnais reste à poser.

`make test` couvre les deux applications d'un coup.

## Architecture

Les deux premiers niveaux du modèle C4. C4 décrit un système à quatre niveaux de
zoom : contexte, conteneurs, composants, code. On s'arrête à deux : les niveaux
3 et 4 redisent ce que le code dit déjà, et se périment à la première refonte,
alors qu'un schéma faux est pire qu'un schéma absent.

### C4 niveau 1, contexte

Qui utilise le système et pour quoi, sans rien dire de la technique.

```mermaid
flowchart TB
    V["Visiteur<br/><i>Personne</i>"]
    A["Administrateur<br/><i>Personne</i>"]
    S["Portail de services<br/><i>Système</i><br/>Publie un catalogue de<br/>prestations et permet<br/>de l'alimenter"]

    V -->|"Consulte le catalogue publié"| S
    A -->|"Crée et publie des services"| S
```

Aucun système tiers : ni paiement, ni messagerie, ni annuaire. Le portail est
autonome, et le schéma le dit plutôt que de laisser imaginer le contraire.

### C4 niveau 2, conteneurs

Les unités déployables séparément, avec la technologie et le protocole sur
chaque échange.

```mermaid
flowchart TB
    V["Visiteur"]
    A["Administrateur"]

    subgraph portail["Portail de services"]
        F["Front<br/><i>React, TypeScript, Vite</i><br/>Catalogue et fiches"]
        API["API<br/><i>Go, chi</i><br/>Lecture et écriture<br/>du catalogue"]
        DB[("Base de données<br/><i>PostgreSQL 17</i><br/>Services, migrations")]
        R[("Cache<br/><i>Redis 7</i><br/>Listing, TTL 30 s")]
    end

    V -->|"HTTPS"| F
    A -->|"HTTPS"| F
    F -->|"REST, JSON sur HTTP<br/>/api/services"| API
    API -->|"SQL sur TCP 5432<br/>pgx, requêtes paramétrées"| DB
    API -.->|"RESP sur TCP 6379<br/>go-redis, optionnel"| R
```

Le trait plein est une dépendance dure, le pointillé une dépendance optionnelle :
si Redis ne répond pas, l'API sert la même réponse depuis PostgreSQL.

En développement le front et l'API sont deux processus, mais une seule origine
pour le navigateur : Vite proxifie `/api` vers l'API.

### Trajet d'une requête

Ce diagramme ne fait pas partie du C4, c'est une séquence. Il montre ce que les
schémas ci-dessus ne peuvent pas montrer : l'ordre des opérations.

```mermaid
sequenceDiagram
    participant C as Client
    participant M as Middlewares
    participant H as Handler
    participant K as Cache Redis
    participant R as Repository
    participant DB as PostgreSQL

    C->>M: GET /api/services
    M->>H: RequestID, RealIP, Logger, Recoverer
    H->>H: lecture et bornage des paramètres
    H->>K: HGET services:list, clé complète
    alt entrée présente
        K-->>H: réponse en cache
    else absente, ou Redis muet
        H->>R: List(ctx, onlyPublished, limit, offset)
        R->>DB: SELECT avec requête paramétrée
        DB-->>R: lignes
        R-->>H: []Service
        H->>K: HSET puis EXPIRE NX, 30 s
    end
    H-->>C: 200 avec la liste en JSON
```

## Choix techniques

**chi plutôt que Gin.** chi reste compatible avec `net/http`, un handler chi est un
`http.HandlerFunc` ordinaire. Rien n'est masqué et on peut en sortir sans réécrire
les handlers.

**Pas d'ORM.** Le SQL est écrit à la main avec pgx. Les requêtes sont paramétrées :
le texte SQL et les valeurs partent séparément, ce qui rend l'injection impossible.
Les colonnes sont listées explicitement plutôt qu'un `SELECT *`, pour qu'un ajout
de colonne ne casse pas le scan.

**Migrations embarquées.** Les fichiers SQL sont inclus dans le binaire avec
`go:embed` et appliqués au démarrage, chacune dans une transaction. Une table de
suivi évite de les rejouer. Le déploiement se réduit à un seul exécutable.

**Prix en centimes.** Stockés en entier, jamais en flottant. `0.1 + 0.2` ne vaut pas
`0.3` en virgule flottante, ce qui produit des erreurs d'arrondi sur des montants.

**Contraintes côté base.** Slug unique, `CHECK` sur la gamme et la durée. La
validation applicative donne des messages clairs, la base reste le dernier rempart
si un bug passe au travers.

**Arrêt propre.** Sur `SIGTERM`, le serveur cesse d'accepter de nouvelles requêtes
et laisse dix secondes à celles en cours pour finir, plutôt que de les couper.

**Pool de connexions borné.** Ouvrir une connexion PostgreSQL coûte cher et le
serveur en accepte un nombre limité. Le pool est plafonné pour ne pas le saturer
sous charge.

**Cache sur le listing, pas sur le détail.** Le listing est la page d'accueil :
une seule réponse sert tous les visiteurs, le taux de réutilisation est élevé. Le
détail est réparti sur autant de slugs qu'il y a de services, chaque entrée
serait lue rarement et occuperait de la mémoire pour rien. Cacher là où ça ne
rapporte pas ajoute un chemin d'invalidation à maintenir sans gain mesurable.

**La clé contient tous les paramètres.** `published`, `limit` et `offset`
changent le résultat, donc les trois entrent dans la clé. En oublier un ferait
servir la page 2 à qui demande la page 1, un bug bien pire que l'absence de
cache. Les valeurs sont bornées par `service.Page` avant de construire la clé et
avant la requête SQL, pour que `limit=0` et `limit=20`, qui donnent le même
résultat, partagent la même entrée.

**TTL court plutôt qu'invalidation fine.** Trente secondes, plus une purge à la
création. Suivre précisément quelles pages une écriture invalide demande de
rejouer la logique de tri et de pagination dans le cache : beaucoup de code, et
un décalage silencieux le jour où la requête change. Le TTL borne l'erreur sans
rien à maintenir. Les entrées vivent dans un seul hash Redis, ce qui rend la
purge atomique et en une commande, sans parcourir l'espace de clés.

**Une panne de Redis ne fait pas tomber l'API.** Le cache est une optimisation,
pas une dépendance. `cache.New` ne se connecte pas au démarrage, chaque commande
est bornée à 200 ms, et toute erreur est journalisée puis ignorée : la réponse
vient de PostgreSQL. Une URL vide donne un cache désactivé, ce qui permet de
déployer sans Redis et de tester sans. Le prix à payer est visible : Redis
injoignable ajoute le délai d'attente à chaque requête, faute de disjoncteur qui
cesserait d'essayer après une série d'échecs.

## À venir

Front : formulaire d'administration avec validation. Authentification et pages
réservées. Déploiement.
