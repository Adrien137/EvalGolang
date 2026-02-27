Projet Go réalisé par Adrien Baldocchi à partir du fichier suivant : https://docs.google.com/document/d/1REj3tFOesvtn8_QxczPVbtlOIR01K8lyiPKu_jL-4xc/edit?tab=t.0

Ce projet est une application en ligne de commande écrite en Go qui permet de réaliser les actions suivantes :

- Analyser des fichiers texte
- Analyser plusieurs fichiers dans un dossier
- Télécharger et analyser une page Wikipédia
- Gérer des processus système
- Appliquer des opérations de sécurité sur des fichiers

Technologies utilisées :

- Go (Golang)
- encoding/json vers la lecture du fichier config
- net/http pour les requêtes web
- github.com/PuerkitoBio/goquery  parsing HTML façon jQuery
- os/exec pour l'exécution de commandes système
- syscall pour gestion des permissions (Windows & Unix)

Librairie externe utilisée :
- PuerkitoBio – goquery
- goquery

Lien vers le GitHub permettant l'implémentation de goquery pour analyser une page Wikipédia :
- go get github.com/PuerkitoBio/goquery

---------------------------------------------------

Le programme utilise un fichier config.json qui permet de prédéfinir les options par defaults.

Exemple :
{
  "default_file": "data/input.txt",
  "base_dir": "data",
  "out_dir": "out",
  "default_ext": ".txt"
}

Si le fichier rentrée par l'utilisateur n’est pas trouvé lors des analyses, alors les valeurs par défaut configuré dans ce fichier json sont utilisées.
Voici ce que donne rend le script une fois lancé : 

---------------------------------------

MENU PRINCIPAL
1 - Analyse fichier
2 - Analyse multi-fichiers
3 - Analyse page wikipedia
4 - ProcessOps
5 - SecureOps
6 - Quitter

--------------------------------------

1) Analyse de fichier (Choix A)

Permet de :
Lire un fichier texte (le chemin par default est depuis le dossier fileops/
Afficher : la Taille, la date de modification, le Nombre de lignes, le Nombre de mots (hors nombres), la Longueur moyenne des mots et un Filtrage par mot-clé

Générer :
- filtered.txt
- filtered_not.txt
- head.txt
- tail.txt

Concepts appris :

- bufio.Scanner
- os.Stat
- strings.Fields
- slices
- gestion d’erreurs

----------------------------------------

2) Analyse multi-fichiers (Choix B)

Parcourt un dossier et analyse tous les fichiers .txt. pour générer :

Un report.txt correspondant au nombre de lignes par fichier
Un index.txt regroupant la taille + la date de modification
et un merged.txt qui correspond à la fusion de tous les fichiers

Concepts appris :
- filepath.Walk
- io.Copy
- manipulation de chemins

-----------------------------------------

3) Analyse Wikipédia (Choix C)

Télécharge une page Wikipédia (version française).
Exemple : Pokémon
URL générée : https://fr.wikipedia.org/wiki/Pokémon
Site utilisé : Wikipédia
Fonctionnement :
- Requête HTTP
- Parsing HTML avec goquery
- Extraction des balises <p>
- Calcul statistiques
- Filtrage par mot-clé
- Génération d’un fichier : wiki_Pokémon.txt

Concepts appris :

- http.NewRequest
- User-Agent ( pour ne pas être bloquer par wikipedia lors du téléchargement de la page )
- parsing DOM
- gestion des réponses HTTP

------------------------------------------

4) ProcessOps (Choix D)
Permet de :

- Lister les processus
Pour Windows on utilise tasklist
Pour macOS on utilise ps -Ao pid,comm

- Filtrer un processus par nom afin d'avoir son PID
Par exemple, on peut écrit Discord pour ensuite voir le PID de discord

- Kill sécurisé d’un processus
Vérifie que le PID existe bien ( renvoie un message d'erreur di il n'existe pas )
Demande la confirmation avant de supprimer
Pour Windows on utilise taskkill
Pour macOS/Linux on utilise kill -9

Concepts appris :

- os/exec
- différences Windows / Unix
- runtime.GOOS

------------------------------------------

5) SecureOps (Choix E)

Fonctionnalités de SécureOps :

- Création d’un fichier .lock du fichier verrouiller
- Modification des permissions sur un fichier/dossier
- Compatible Windows & Unix
- Journalisation dans un fichier nommée audit.log repertoriant automatiquement tout les fichiers/dossiers lock et unlock ( fichier se reconstruisant si supprimé lors du lancement du programme de vérrouillage/dévérouillage)

Menu SecureOps :
- Verrouiller un fichier (.lock) (crée le fichier en .lock dans le dossier /out)
- Déverrouiller (déverouille tout fichier en .lock puis le supprime)
- Mettre en lecture seule
- Retirer lecture seule
- Vérifier permissions

Concepts appris :
- os.Chmod
- syscall
- bitwise operations permettant de modifier (0222)
- gestion des attributs Windows

---------------------------------------------------

La structure du projet à été réalisé ainsi ( tout les fichiers crée par le programme vont ou seront crée dans /out mais se mettent a jour automatiquement lors des executions du script

/fileops = Eval.go / config.json / data / out

------------------------------------------------------

Pour lancer le programme, il suffit d'ouvrir l'invite de commande puis de se déplacer à l'endroit ou ce trouve le fichier /fileops (assurer vous d'avoir go d'installer, sinon cela ne fonctionnera pas, voici le site officiel pour télécharger go : https://go.dev/dl/ ) puis lancer la commande permettant de lancer le script : go run Eval.go

Cela fonctionne aussi avec un fichier config personnalisée, voici la commande pour le faire :
go run Eval.go -config autre_config.json
