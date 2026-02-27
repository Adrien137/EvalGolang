Projet Go réaliser par Adrien Baldocchi à partie du fichier suivant : https://docs.google.com/document/d/1REj3tFOesvtn8_QxczPVbtlOIR01K8lyiPKu_jL-4xc/edit?tab=t.0

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
- os/exec → exécution de commandes système
- syscall → gestion des permissions (Windows & Unix)

Librairie externe utilisée :
- PuerkitoBio – goquery
- goquery

Lien vers le GitHub permettant l'implémentation de goquery pour analyser une page Wikipédia :
- go get github.com/PuerkitoBio/goquery

Le programme utilise un fichier config.json qui permet de prédéfinir les options par defaults si rien n'est rentrer dans le terminal de commande lors des choix.

Exemple :
{
  "default_file": "data/input.txt",
  "base_dir": "data",
  "out_dir": "out",
  "default_ext": ".txt"
}

Si le fichier n’est pas trouvé alors des valeurs par défaut sont utilisées.
