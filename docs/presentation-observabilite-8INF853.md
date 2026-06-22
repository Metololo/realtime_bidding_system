# Présentation orale II — 8INF853 Architecture des applications d'entreprise

Sujet: Observabilité d'un système d'enchères temps réel avec la stack LGTM
Projet support: Realtime Bidding System
Durée cible: 10 à 12 minutes
Angle: définir le problème architectural, proposer une solution observable, et montrer comment l'intégrer sans casser la Clean Architecture / architecture hexagonale.

---

## Slide 1 — Titre

### Contenu à afficher

Observabilité d'une application d'entreprise distribuée

Cas d'étude: Realtime Bidding System

Stack proposée: LGTM
- Loki: logs
- Grafana: visualisation
- Tempo: traces distribuées
- Mimir / Prometheus: métriques

8INF853 — Architecture des applications d'entreprise — Été 2026

### Visuel suggéré

Schéma simple: plusieurs services à gauche, flèches vers une plateforme d'observabilité à droite.

### Script oral

Bonjour, aujourd'hui je présente mon sujet de projet pour le cours: l'observabilité dans une architecture d'application d'entreprise distribuée. Je vais l'appliquer à mon projet Realtime Bidding System, un système de simulation d'enchères en temps réel. Ce projet est intéressant pour ce sujet parce qu'il combine plusieurs éléments typiques des applications d'entreprise modernes: plusieurs services, communication réseau, messages asynchrones, contraintes de performance, et une architecture propre avec séparation entre domaine, application et infrastructure.

L'objectif n'est pas seulement d'ajouter des dashboards. L'objectif est de résoudre un vrai problème architectural: comment comprendre, diagnostiquer et mesurer un système distribué quand une requête traverse plusieurs composants très rapidement.

---

## Slide 2 — Contexte du projet

### Contenu à afficher

Realtime Bidding System

But du système:
- simuler une plateforme d'enchères très courte durée
- chaque enchère dure environ 100 ms
- les offres en retard sont rejetées
- objectif: haut débit, faible latence

Composants principaux:
- seller / auction generator
- auction engine
- bidder services
- NATS pour les événements
- HTTP/gRPC pour les commandes

### Visuel suggéré

Diagramme de flux:
Seller -> Auction Engine -> NATS -> Bidders -> Auction Engine

### Script oral

Le projet simule une plateforme d'enchères en temps réel. Un service vendeur ou générateur crée des enchères, le moteur d'enchères garde l'état actif en mémoire, publie des événements, et des services bidders réagissent à ces événements pour placer des offres.

Ce système est volontairement orienté haute fréquence: les enchères sont très courtes, environ 100 millisecondes, et le système vise un volume élevé d'enchères et d'offres. Donc une erreur ou une latence de quelques millisecondes peut changer le résultat d'une enchère.

C'est justement ce qui rend l'observabilité pertinente: dans un système simple, des logs peuvent suffire. Ici, il faut comprendre le comportement global du système, pas seulement l'état d'un service isolé.

---

## Slide 3 — Problème architectural

### Contenu à afficher

Problème:

Dans un système distribué à faible latence, il devient difficile de répondre à ces questions:

- Pourquoi une offre a été rejetée?
- Est-ce un problème de logique métier, de latence réseau ou de surcharge?
- Quel service ralentit le chemin critique?
- Est-ce que NATS, HTTP/gRPC ou le moteur d'enchères est le goulot?
- Le système respecte-t-il ses objectifs de 100 ms?

Conséquence:
Sans observabilité, on débogue à l'aveugle.

### Visuel suggéré

Image d'une requête qui traverse plusieurs boîtes, avec un point d'interrogation au milieu.

### Script oral

Le problème que je veux résoudre est le manque de visibilité dans une architecture distribuée. Dans ce projet, une enchère peut être créée, publiée comme événement, consommée par plusieurs bidders, puis retourner vers le moteur d'enchères sous forme d'offres.

Si une offre est rejetée, plusieurs causes sont possibles. Elle peut être arrivée trop tard, le montant peut être insuffisant, le bidder peut avoir déjà participé, le moteur peut être surchargé, ou il peut y avoir une latence dans le réseau ou dans NATS.

Le problème architectural est donc le suivant: comment concevoir l'application pour qu'elle soit observable dès le départ, c'est-à-dire capable d'expliquer son comportement en production ou en simulation.

---

## Slide 4 — Pourquoi ce projet est un bon cas d'étude

### Contenu à afficher

Ce projet est pertinent pour l'observabilité parce qu'il contient:

1. Beaucoup d'événements
   - auction.created
   - auction.closed
   - bid submitted / rejected / accepted

2. Plusieurs styles de communication
   - commandes synchrones HTTP/gRPC
   - événements asynchrones NATS

3. Contraintes fortes
   - latence basse
   - enchères de 100 ms
   - rejet rapide en surcharge

4. Architecture propre
   - domaine
   - application
   - adaptateurs infrastructure

### Visuel suggéré

Grille 2x2: événements, communication, performance, architecture.

### Script oral

Ce projet est un bon cas d'étude parce qu'il n'est pas seulement une application CRUD. Il y a un trafic important, plusieurs composants qui communiquent, et des décisions métier qui dépendent du temps.

Il y a aussi un aspect intéressant pour le cours: l'architecture est organisée autour d'une logique de Clean Architecture ou architecture hexagonale. Le domaine contient les règles d'enchères, l'application orchestre les cas d'utilisation, et l'infrastructure contient les adaptateurs comme HTTP, gRPC, NATS ou la persistance.

Donc je peux montrer comment ajouter l'observabilité comme une capacité transversale, sans mélanger la logique métier avec les outils techniques.

---

## Slide 5 — Objectifs d'observabilité

### Contenu à afficher

Objectifs de la solution:

- Voir l'état du système en temps réel
- Diagnostiquer rapidement une enchère problématique
- Mesurer la latence de bout en bout
- Corréler logs, traces et métriques
- Détecter les goulots d'étranglement
- Comparer les performances entre exécutions

Principe clé:
Chaque événement important doit être lié à un contexte: auction_id, bidder_id, trace_id.

### Visuel suggéré

Triangle des trois signaux: Logs, Metrics, Traces, avec Grafana au centre.

### Script oral

L'objectif de la solution est de rendre le système explicable. Pour cela, je vais utiliser les trois signaux classiques de l'observabilité: les logs, les métriques et les traces.

Les logs expliquent ce qui s'est passé. Les métriques indiquent la santé et la performance globale. Les traces montrent le parcours d'une opération à travers plusieurs services.

Le point central est la corrélation. Si j'ai un auction_id ou un trace_id, je dois pouvoir retrouver les logs associés, la trace distribuée et les métriques pertinentes pour comprendre ce qui s'est passé.

---

## Slide 6 — Stack LGTM proposée

### Contenu à afficher

Stack LGTM:

L — Loki
- stockage et recherche des logs
- logs structurés JSON

G — Grafana
- dashboards
- exploration
- alertes

T — Tempo
- traces distribuées
- parcours d'une enchère entre services

M — Mimir ou Prometheus
- métriques applicatives et système
- latence, débit, erreurs, saturation

Collecte recommandée:
- OpenTelemetry Collector

### Visuel suggéré

Pipeline: Services -> OpenTelemetry Collector -> Loki / Tempo / Prometheus-Mimir -> Grafana.

### Script oral

La stack proposée est LGTM. Dans ce contexte, Loki sert aux logs, Grafana sert à visualiser et explorer les données, Tempo stocke les traces distribuées, et Mimir ou Prometheus gère les métriques.

Pour éviter de coupler chaque service directement à chaque outil, j'utiliserais OpenTelemetry Collector. Les services instrumentés envoient leurs données au collector, et le collector les route vers les bons backends.

Ce choix est intéressant architecturalement parce qu'il sépare l'instrumentation de l'application et l'infrastructure d'observabilité. Si plus tard on change Tempo ou Loki, l'application n'a pas besoin d'être fortement modifiée.

---

## Slide 7 — Intégration avec l'architecture hexagonale

### Contenu à afficher

Principe:
L'observabilité ne doit pas polluer le domaine.

Domaine:
- règles d'enchères
- validation des offres
- choix du gagnant

Application:
- cas d'utilisation
- création d'enchère
- soumission d'offre
- fermeture d'enchère

Infrastructure / adaptateurs:
- HTTP/gRPC instrumentation
- NATS instrumentation
- logs structurés
- métriques Prometheus
- OpenTelemetry traces

### Visuel suggéré

Hexagone: domaine au centre, application autour, adaptateurs autour. Observabilité comme couche transversale dans les adaptateurs et cas d'utilisation.

### Script oral

Un enjeu important est de ne pas transformer l'observabilité en dépendance métier. Le domaine ne devrait pas connaître Loki, Grafana, Tempo ou Prometheus.

Par exemple, la règle qui décide si une offre est valide doit rester dans le domaine. Par contre, l'adaptateur HTTP ou gRPC peut mesurer la durée d'une requête, ajouter un trace_id, et produire des logs structurés.

L'application peut aussi exposer des points d'observation au niveau des cas d'utilisation: création d'enchère, soumission d'offre, fermeture d'enchère. Mais les détails techniques restent dans l'infrastructure.

C'est là que le projet rejoint directement le cours: l'observabilité devient une préoccupation architecturale, pas seulement un ajout technique.

---

## Slide 8 — Traces distribuées: suivre une enchère complète

### Contenu à afficher

Trace cible: cycle de vie d'une enchère

1. Seller crée une enchère
2. Auction Engine valide et sauvegarde l'état actif
3. Auction Engine publie auction.created dans NATS
4. Bidder reçoit l'événement
5. Bidder soumet une offre
6. Auction Engine valide l'offre
7. Auction Engine ferme l'enchère
8. Événement auction.closed publié

Données de corrélation:
- trace_id
- auction_id
- bidder_id
- event_id

### Visuel suggéré

Waterfall de trace distribuée avec spans: seller, auction-engine, nats publish, bidder, submit bid, close auction.

### Script oral

Les traces distribuées sont probablement l'élément le plus important dans ce projet. Une enchère traverse plusieurs composants en très peu de temps. Avec Tempo, je veux pouvoir reconstruire le chemin complet.

Par exemple, une trace pourrait commencer au moment où le seller crée l'enchère. Ensuite, on verrait le traitement dans le moteur, la publication dans NATS, la réception par un bidder, puis la soumission d'une offre au moteur.

Chaque span donne une durée. On peut donc voir si le problème vient du seller, du moteur, du broker NATS ou du bidder. Cette information est difficile à obtenir avec des logs isolés.

---

## Slide 9 — Métriques: mesurer la santé et la performance

### Contenu à afficher

Métriques applicatives proposées:

Débit:
- auctions_created_total
- auctions_closed_total
- bids_submitted_total

Latence:
- auction_creation_duration_seconds
- bid_processing_duration_seconds
- auction_lifecycle_duration_seconds

Erreurs / rejets:
- bids_rejected_total{reason}
- auction_create_failed_total{reason}

Saturation:
- active_auctions
- nats_publish_errors_total
- goroutines / mémoire / CPU

### Visuel suggéré

Mock dashboard Grafana: débit, p95 latence, rejets par raison, enchères actives.

### Script oral

Les métriques servent à comprendre l'état global du système. Contrairement aux traces, qui expliquent une opération précise, les métriques permettent de voir les tendances et les anomalies.

Je suivrais par exemple le nombre d'enchères créées, le nombre d'offres reçues, les rejets par raison, la latence de traitement des offres, et le nombre d'enchères actives.

Ces métriques sont importantes parce que le système a des contraintes explicites: les enchères sont courtes et le système doit rejeter rapidement en cas de surcharge. Les métriques permettent de vérifier si cette promesse est respectée.

---

## Sl`ide 10 — Logs structurés: expliquer les décisions métier

### Contenu à afficher

Logs structurés JSON avec champs constants:

- timestamp
- level
- service
- operation
- trace_id
- auction_id
- bidder_id
- event_id
- reason
- latency_ms

Exemples de logs utiles:
- auction created
- bid accepted
- bid rejected: auction_expired
- auction closed
- nats publish failed

### Visuel suggéré

Exemple de log JSON affiché à gauche, recherche dans Grafana/Loki à droite.

### Script oral

Les logs restent importants, mais ils doivent être structurés. Dans ce projet, un log texte comme "bid failed" n'est pas suffisant. Il faut savoir quelle enchère, quel bidder, quelle raison de rejet, et dans quelle trace.

Les logs sont particulièrement utiles pour expliquer les décisions métier: pourquoi une offre a été rejetée, pourquoi une enchère n'a pas eu de gagnant, ou pourquoi un événement n'a pas été publié.

Avec Loki, je peux filtrer par auction_id ou par reason. Et si le log contient le trace_id, je peux passer du log à la trace distribuée correspondante.

---

## Slide 11 — Exemple de scénario de diagnostic

### Contenu à afficher

Incident simulé:
"Un bidder perd trop souvent malgré des offres valides."

Démarche avec LGTM:

1. Grafana: hausse de bids_rejected_total{reason="auction_expired"}
2. Tempo: traces montrent un délai entre NATS et bidder
3. Loki: logs confirment réception tardive des événements
4. Métriques système: CPU ou goroutines élevés sur bidder
5. Conclusion: goulot dans bidder ou consommation NATS

Résultat:
Diagnostic basé sur des données corrélées.

### Visuel suggéré

Chaîne d'enquête: métrique -> trace -> log -> cause racine.

### Script oral

Voici un exemple concret. Supposons qu'un bidder perd souvent ou que ses offres sont rejetées comme expirées. Sans observabilité, on pourrait penser que la logique métier est incorrecte.

Avec la stack LGTM, je commence par les métriques et je vois que les rejets pour auction_expired augmentent. Ensuite, je regarde une trace dans Tempo et je vois que le délai se produit entre la publication NATS et le traitement par le bidder. Puis, dans Loki, les logs confirment que le bidder reçoit certains événements trop tard.

À ce moment-là, le problème n'est probablement pas la règle métier. C'est plutôt un problème de consommation, de charge ou de concurrence dans le bidder. L'observabilité permet donc de passer d'un symptôme à une cause probable.

---

## Slide 12 — Plan d'application technique

### Contenu à afficher

Étapes d'implémentation:

1. Ajouter OpenTelemetry aux services Go
2. Instrumenter HTTP/gRPC
3. Propager le contexte dans NATS
4. Ajouter métriques Prometheus
5. Standardiser les logs JSON avec slog
6. Déployer LGTM avec Docker Compose
7. Créer dashboards Grafana
8. Tester avec un scénario de charge

Livrable technique:
Une version observable du système d'enchères.

### Visuel suggéré

Roadmap en 8 étapes.

### Script oral

Pour la partie technique du projet, je procéderais par étapes. Premièrement, j'ajoute OpenTelemetry aux services Go. Ensuite, j'instrumente les entrées HTTP ou gRPC pour créer les traces automatiquement.

Pour NATS, il faut propager le contexte de trace dans les headers des messages. C'est un point important parce que les communications asynchrones cassent souvent la continuité des traces si on ne propage pas explicitement le contexte.

Ensuite, j'ajoute des métriques Prometheus et je standardise les logs JSON avec slog. Finalement, je déploie Loki, Grafana, Tempo et Prometheus ou Mimir avec Docker Compose, puis je crée des dashboards pour observer le système pendant une simulation de charge.

---

## Slide 13 — Apport pour l'architecture d'entreprise

### Contenu à afficher

Valeur architecturale:

- meilleure maintenabilité
- diagnostic plus rapide
- validation des exigences non fonctionnelles
- séparation claire entre domaine et infrastructure
- meilleure compréhension des communications interservices
- base pour alertes et SLO

Lien avec le cours:
L'observabilité devient une qualité architecturale mesurable.

### Visuel suggéré

Avant / après:
Avant: services opaques
Après: services observables avec signaux corrélés

### Script oral

L'intérêt pour l'architecture d'entreprise est que l'observabilité permet de valider des exigences non fonctionnelles. On ne se contente pas de dire que le système est rapide ou résilient: on le mesure.

Elle améliore aussi la maintenabilité. Quand le système évolue, les traces, les logs et les métriques permettent de vérifier si un changement a dégradé la performance ou introduit des erreurs.

Dans une architecture hexagonale, cela montre aussi qu'on peut ajouter des capacités transversales tout en gardant une séparation propre entre le domaine et l'infrastructure.

---

## Slide 14 — Conclusion

### Contenu à afficher

Conclusion:

Problème:
Un système d'enchères distribué à faible latence est difficile à comprendre sans observabilité.

Solution:
Mettre en place une architecture observable avec LGTM et OpenTelemetry.

Contribution du projet:
- cas réaliste de trafic élevé
- communications synchrones et asynchrones
- architecture hexagonale
- métriques, logs et traces corrélés

Phrase finale:
Rendre le système observable, c'est rendre son architecture vérifiable.

### Visuel suggéré

Schéma final: architecture applicative + couche d'observabilité + dashboards.

### Script oral

Pour conclure, le problème que je veux traiter est la difficulté de comprendre un système distribué à faible latence. Mon projet d'enchères temps réel est un bon support parce qu'il génère beaucoup d'événements, implique plusieurs services et contient des contraintes de performance fortes.

La solution proposée est d'intégrer une stack d'observabilité LGTM avec OpenTelemetry. Cela permet de corréler logs, métriques et traces autour d'identifiants comme auction_id, bidder_id et trace_id.

L'apport principal est architectural: l'observabilité devient une façon de vérifier que le système respecte ses objectifs et que les choix d'architecture fonctionnent réellement.

---

## Questions possibles du professeur et réponses rapides

### Question: Pourquoi ne pas seulement utiliser des logs?

Réponse:
Les logs expliquent des événements locaux, mais ils ne montrent pas facilement le chemin complet d'une requête entre plusieurs services. Les traces distribuées complètent les logs en montrant le parcours et les durées entre composants.

### Question: Pourquoi LGTM plutôt qu'une autre stack?

Réponse:
LGTM est cohérente parce qu'elle couvre les trois signaux d'observabilité avec des outils intégrés autour de Grafana. Elle est aussi facile à déployer localement avec Docker Compose, ce qui convient bien à un projet de cours.

### Question: Est-ce que l'observabilité va ralentir le système?

Réponse:
Oui, elle ajoute un coût, mais il peut être contrôlé avec l'échantillonnage des traces, des métriques agrégées et des logs structurés au bon niveau. L'objectif est de trouver un équilibre entre visibilité et performance.

### Question: Où placer l'observabilité dans une architecture hexagonale?

Réponse:
Principalement dans les adaptateurs et les cas d'utilisation, pas dans le domaine pur. Le domaine garde les règles métier; l'infrastructure gère l'instrumentation technique.

### Question: Comment prouver que la solution fonctionne?

Réponse:
En exécutant une simulation de charge, puis en montrant dans Grafana les métriques de débit, les traces d'une enchère complète et les logs corrélés par auction_id ou trace_id.
