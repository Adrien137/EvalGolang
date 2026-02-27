package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Config struct pour JSON
type Config struct {
	DefaultFile string `json:"default_file"`
	BaseDir     string `json:"base_dir"`
	OutDir      string `json:"out_dir"`
	DefaultExt  string `json:"default_ext"`
}

func main() {
	// Flag JSON config
	configPath := flag.String("config", "config.json", "Chemin vers config JSON")
	flag.Parse()

	cfg := loadConfig(*configPath)
	reader := bufio.NewReader(os.Stdin)

	// Création du dossier out si inexistant
	os.MkdirAll(cfg.OutDir, os.ModePerm)

	for {
		fmt.Println("\n===== MENU =====")
		fmt.Println()
		fmt.Println("1 - Choix A (Analyse fichier)")
		fmt.Println("2 - Choix B (Analyse multi-fichiers)")
		fmt.Println("3 - Choix C (Analyse page wikipedia)")
		fmt.Println("4 - Choix D (ProcessOps)")
		fmt.Println("5 - Choix E (SecureOps)")
		fmt.Println("6 - QUITTER")
		fmt.Println()
		fmt.Print("Choix : ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			choixA(cfg, reader)
		case "2":
			choixB(cfg, reader)
		case "3":
			choixWiki(cfg, reader)
		case "4":
			choixProcessOps(reader)
		case "5":
			secureOpsMenu(cfg, reader)
		case "6":
			fmt.Println("Fin du programme.")
			return
		default:
			fmt.Println("Choix invalide.")
		}
	}
}

// ------- 10/ 20 : analyse de fichier --------

// Fonction pour charger config.json
func loadConfig(path string) Config {
	cfg := Config{
		DefaultFile: "data/input.txt",
		BaseDir:     "data",
		OutDir:      "out",
		DefaultExt:  ".txt",
	}

	// Lire le fichier config.json
	file, err := os.ReadFile(path)
	if err != nil {
		fmt.Println("config.json non trouvé, valeurs par défaut utilisées.")
		return cfg
	}

	// Parse le fichier JSON
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		fmt.Println("Erreur parsing config JSON, valeurs par défaut utilisées.")
		return cfg
	}
	return cfg
}

// Fonction pour demander un chemin de fichier/dossier avec une valeur par défaut
func askPath(reader *bufio.Reader, def string) string {
	fmt.Printf("Choisi le fichier par default input.txt si valeur vide = %s) : ", def)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return def
	}
	return input
}

// Choix A
func choixA(cfg Config, reader *bufio.Reader) {
	path := askPath(reader, cfg.DefaultFile)

	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		fmt.Println("Fichier invalide.")
		return
	}

	// Afficher les infos du fichier
	fmt.Println("Taille :", info.Size(), "bytes")
	fmt.Println("Modifié :", info.ModTime().Format(time.RFC3339))

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Erreur ouverture fichier.")
		return
	}
	defer file.Close()

	// Lire les lignes du fichier et les stocker dans un slice
	var lines []string
	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line != "" {
			lines = append(lines, line)
		}
	}
	fmt.Println("Nombre de lignes :", len(lines))

	// Stats des mots (en ignorant les valeurs numériques)
	totalWords := 0
	totalLen := 0
	for _, l := range lines {
		for _, w := range strings.Fields(l) {
			if _, err := strconv.Atoi(w); err != nil {
				totalWords++
				totalLen += len(w)
			}
		}
	}

	// On affiche les stats si on a au moins un mot
	if totalWords > 0 {
		fmt.Println("Nombre de mots :", totalWords)
		fmt.Println("Longueur moyenne :", totalLen/totalWords)
	}

	// On rentre le Mot-clé
	fmt.Print("Mot-clé : ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)

	// Créer le dossier out si inexistant
	os.MkdirAll(cfg.OutDir, os.ModePerm)

	// Fichiers de sortie
	fYes, _ := os.Create(filepath.Join(cfg.OutDir, "filtered.txt"))
	defer fYes.Close()
	fNo, _ := os.Create(filepath.Join(cfg.OutDir, "filtered_not.txt"))
	defer fNo.Close()

	// Filtrer les lignes et écrire dans les fichiers de sortie
	count := 0
	for _, l := range lines {
		lTrim := strings.TrimSpace(l) // supprime espaces début/fin
		if lTrim == "" {
			continue // ignore les lignes vides
		}

		// Si mot-clé non vide, filtrer les lignes
		if keyword != "" {
			if strings.Contains(lTrim, keyword) {
				fYes.WriteString(lTrim + "\n")
				count++
			} else {
				fNo.WriteString(lTrim + "\n")
			}
		} else {
			// mot-clé vide -> toutes les lignes non vides vont dans filtered.txt
			fYes.WriteString(lTrim + "\n")
		}
	}

	fmt.Println("Lignes contenant le mot-clé :", count)
	fmt.Println("Fichiers générés dans", cfg.OutDir)

	// Head / Tail
	fmt.Print("Choix des lignes à garder pour head/tail : ")
	nStr, _ := reader.ReadString('\n')
	n, _ := strconv.Atoi(strings.TrimSpace(nStr))

	if n > len(lines) {
		n = len(lines)
	}

	head := strings.Join(lines[:n], "\n")
	tail := strings.Join(lines[len(lines)-n:], "\n")

	// Écrire head et tail dans des fichiers suivants : head.txt et tail.txt
	os.WriteFile(cfg.OutDir+"/head.txt", []byte(head), 0644)
	os.WriteFile(cfg.OutDir+"/tail.txt", []byte(tail), 0644)

	fmt.Println("Fichiers générés dans", cfg.OutDir)
}

// choix B
func choixB(cfg Config, reader *bufio.Reader) {
	dir := askPath(reader, cfg.BaseDir)
	os.MkdirAll(cfg.OutDir, os.ModePerm)

	// Fichiers de sortie (out)
	report, _ := os.Create(cfg.OutDir + "/report.txt")
	index, _ := os.Create(cfg.OutDir + "/index.txt")
	merged, _ := os.Create(cfg.OutDir + "/merged.txt")

	defer report.Close()
	defer index.Close()
	defer merged.Close()

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// On ne traite que les fichiers avec l'extension par défaut
		if !info.IsDir() && strings.HasSuffix(path, cfg.DefaultExt) {

			index.WriteString(fmt.Sprintf("%s | %d bytes | %s\n",
				path, info.Size(), info.ModTime().Format(time.RFC3339)))

			file, err := os.Open(path)

			// Vérifier l'ouverture du fichier
			if err != nil {
				return nil
			}

			// Compter les lignes
			sc := bufio.NewScanner(file)
			lines := 0
			for sc.Scan() {
				lines++
			}
			report.WriteString(fmt.Sprintf("%s : %d lignes\n", path, lines))

			// Revenir au début du fichier pour la copie
			file.Seek(0, 0)
			_, err = io.Copy(merged, file)
			if err != nil {
				fmt.Println("Erreur copie:", err)
			}
			// Vérifier que le fichier se termine par un '\n'
			merged.WriteString("\n")
		}
		return nil
	})
	fmt.Println("Analyse multi-fichiers terminée.")
}

// ------- 12/ 20 : Page Wikipédia --------
// Téléchargement et analyse d'une page Wikipédia

// Cette fonction :
// 1) Demande un nom d'article
// 2) Télécharge la page HTML depuis Wikipédia
// 3) Extrait le texte des balises <p>
// 4) Calcule des statistiques sur les mots
// 5) Sauvegarde les paragraphes filtrés dans un fichier

func choixWiki(cfg Config, reader *bufio.Reader) {

	// Demande à l'utilisateur le nom exact de l'article Wikipédia
	fmt.Print("Article Wikipédia : ")
	article, _ := reader.ReadString('\n')
	article = strings.TrimSpace(article)

	// Si l'utilisateur n'entre rien alors on arrête
	if article == "" {
		fmt.Println("Article vide, abandon.")
		return
	}

	// Construction de l'URL vers Wikipédia
	url := "https://fr.wikipedia.org/wiki/" + article
	fmt.Println("Téléchargement de :", url)

	// Création d’un client HTTP
	client := &http.Client{}

	// Création d’une requête GET vers l’URL de l’article
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Erreur création requête :", err)
		return
	}

	// Ajout d’un User-Agent pour simuler un navigateur afin de ne pas se faire bloquer( ici mozzila )
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

	// Envoi de la requête HTTP et télécharge tout le HTML de la page Wikipédia
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Erreur téléchargement :", err)
		return
	}
	defer resp.Body.Close()

	// Vérification du code HTTP, 200 = OK et 404 = page inexistante
	if resp.StatusCode != 200 {
		fmt.Println("Erreur HTTP :", resp.Status)
		return
	}

	// Analyse du HTML avec goquery qui va construire un DOM à partir du HTML téléchargé
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Println("Erreur analyse HTML :", err)
		return
	}

	// Le tableau qui va contenir le TEXTE des paragraphes
	var lines []string

	// doc.Find("p") sélectionne TOUTES les balises <p> de la page HTML
	// Cela inclut :
	// - Le texte principal de l'article
	// - Les paragraphes secondaires
	// - Les éventuels paragraphes hors contenu principal
	doc.Find("p").Each(func(i int, s *goquery.Selection) {

		// s.Text() récupère tout le texte contenu dans la balise <p>
		// Voici ce qui est inclus dedans :
		// Le texte normal
		// Le texte des liens <a>
		// Le texte des <span>
		// Les références [1], [2] dans les <sup>
		text := strings.TrimSpace(s.Text())

		// On ignore les paragraphes vides
		if text != "" {
			lines = append(lines, text)
		}
	})

	// Affiche combien de paragraphes ont été extraits
	fmt.Println("Paragraphes extraits :", len(lines))

	// Nombre total de mots détecté Somme des longueurs de tous les mots
	totalWords := 0
	totalLen := 0

	// Pour chaque paragraphe extrait
	for _, l := range lines {

		// strings.Fields va découper la phrase en "mots" en séparant par espaces
		// La ponctuation reste attachée (par exemple : "France.")
		for _, w := range strings.Fields(l) {

			// strconv.Atoi(w) Tente de convertir le mot en entier
			// Si la conversion réussit alors c’est un nombre pur (exemple: "1998")
			// Si la conversion échoue alors ce n’est PAS un entier pur
			if _, err := strconv.Atoi(w); err != nil {

				// On compte uniquement les mots non numériques purs
				totalWords++
				// On ajoute la longueur brute du mot (ponctuation incluse si présente)
				totalLen += len(w)
			}
		}
	}

	// Calcul de la moyenne de longueur
	// Moyenne = somme longueurs / nombre de mots
	if totalWords > 0 {
		fmt.Println("Nombre de mots :", totalWords)
		fmt.Println("Longueur moyenne :", totalLen/totalWords)
	}

	// FILTRAGE PAR MOT-CLÉ
	fmt.Print("Mot-clé pour filtrer (ENTER = aucun) : ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)

	// Création du dossier de sortie si inexistant
	os.MkdirAll(cfg.OutDir, os.ModePerm)

	// Création du fichier :
	outFile := filepath.Join(cfg.OutDir, "wiki_"+article+".txt")
	f, err := os.Create(outFile)
	if err != nil {
		fmt.Println("Erreur création fichier :", err)
		return
	}
	defer f.Close()

	// ÉCRITURE DES DONNÉES
	count := 0
	for _, l := range lines {

		// Si aucun mot-clé alors on écrit tout
		// Sinon, on écrit uniquement les paragraphes contenant le mot-clé
		if keyword == "" || strings.Contains(l, keyword) {
			f.WriteString(l + "\n")

			if keyword != "" {
				count++
			}
		}
	}

	// Résumé final
	fmt.Println("Fichier généré :", outFile)
	if keyword != "" {
		fmt.Println("Lignes contenant le mot-clé :", count)
	}
}

// ------- 14/ 20 : ProcessOps --------
// CHOIX D : ProcessOps (Lister processus, filtrer, kill sécurisé)
func choixProcessOps(reader *bufio.Reader) {
	for {
		fmt.Println("\n-------- ProcessOps --------")
		fmt.Println()
		fmt.Println("1 - Lister les processus")
		fmt.Println("2 - Rechercher / filtrer un processus")
		fmt.Println("3 - Kill sécurisé d'un processus")
		fmt.Println("4 - Retour au menu principal")
		fmt.Println()
		fmt.Print("Choix : ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			listProcesses(reader)
		case "2":
			filterProcesses(reader)
		case "3":
			killProcess(reader)
		case "4":
			return
		default:
			fmt.Println("Choix invalide.")
		}
	}
}

// Cette fonction permet de lister les processus en fonction du système d'exploitation
func listProcesses(reader *bufio.Reader) {
	fmt.Print("Nombre de processus à afficher (entrer un nombre svp par pitié) : ")
	nStr, _ := reader.ReadString('\n')
	nStr = strings.TrimSpace(nStr)
	n := 10 // Valeur par défaut
	if nStr != "" {
		fmt.Sscanf(nStr, "%d", &n)
	}

	osName := runtime.GOOS
	var cmd *exec.Cmd
	// La commande pour lister les processus dépend du système d'exploitation
	if osName == "windows" { //windows = windows
		cmd = exec.Command("tasklist", "/FO", "CSV") //utilisation de tasklist pour Windows, format CSV pour faciliter le parsing
	} else if osName == "darwin" { //darwin = macOS
		cmd = exec.Command("ps", "-Ao", "pid,comm") // sinon, utilisation de ps pour MacOS
	} else {
		fmt.Println("OS non supporté") //OS pas supporté, désolé pour le dérangement
		return
	}

	// Exécuter la commande et récupérer la sortie
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}

	//On traiter la sortie pour n'afficher que les N premiers processus
	lines := strings.Split(string(out), "\n")

	//On ignore la première ligne CSV (en-tête) pour Windows
	if osName == "windows" && len(lines) > 0 {
		lines = lines[1:] //saute la ligne d'en-tête du tasklist
	}

	// Maintenant on limite aux N premiers processus
	if len(lines) > n {
		lines = lines[:n]
	}

	// Afficher les processus
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Si on est sur Windows, on doit parser le CSV pour extraire le nom du processus
		if osName == "windows" {
			fields := strings.Split(line, ",")
			if len(fields) >= 2 {
				pid := strings.Trim(fields[1], "\"")
				name := strings.Trim(fields[0], "\"")
				fmt.Println(pid, "\t", name)
			}
		} else {
			fmt.Println(line)
		}
	}
}

// Cette fonction permet de filtrer les processus en fonction d'un mot-clé dans leur nom
func filterProcesses(reader *bufio.Reader) {
	fmt.Print("Mot à rechercher dans le nom : ")
	keyword, _ := reader.ReadString('\n')
	keyword = strings.TrimSpace(keyword)
	if keyword == "" {
		fmt.Println("Mot vide, abandon.")
		return
	}

	// La commande pour lister les processus dépend du système d'exploitation
	osName := runtime.GOOS
	var cmd *exec.Cmd
	if osName == "windows" {
		cmd = exec.Command("tasklist", "/FO", "CSV")
	} else if osName == "darwin" {
		cmd = exec.Command("ps", "-Ao", "pid,comm")
	} else {
		fmt.Println("OS non supporté")
		return
	}

	// Exécuter la commande et récupérer la sortie
	out, err := cmd.Output()
	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}

	// Traiter la sortie et filtrer les processus contenant le mot-clé
	lines := strings.Split(string(out), "\n")
	fmt.Println("PID\tNom")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		} // Si on est sur Windows, on doit parser le CSV pour extraire le nom du processus
		if osName == "windows" {
			fields := strings.Split(line, ",")
			if len(fields) >= 2 {
				pid := strings.Trim(fields[1], "\"")
				name := strings.Trim(fields[0], "\"")
				if strings.Contains(strings.ToLower(name), strings.ToLower(keyword)) {
					fmt.Println(pid, "\t", name)
				}
			}
		} else { // Sur MacOS/Linux, la ligne contient déjà le PID et le nom, on peut filtrer directement
			if strings.Contains(strings.ToLower(line), strings.ToLower(keyword)) {
				fmt.Println(line)
			}
		}
	}
}

// Cette fonction permet de tuer un processus de manière sécurisée
func killProcess(reader *bufio.Reader) {
	fmt.Print("PID à tuer : ")
	pidStr, _ := reader.ReadString('\n')
	pidStr = strings.TrimSpace(pidStr)
	if pidStr == "" {
		fmt.Println("PID vide, abandon.")
		return
	}

	// Vérifier que le PID existe avant de tenter de le tuer
	osName := runtime.GOOS
	var cmdCheck *exec.Cmd
	if osName == "windows" {
		cmdCheck = exec.Command("tasklist", "/FI", "PID eq "+pidStr, "/FO", "CSV", "/NH") // /NH supprime l’en-tête
	} else {
		cmdCheck = exec.Command("ps", "-p", pidStr, "-o", "pid,comm")
	}

	// Exécuter la commande de vérification et récupérer la sortie
	out, err := cmdCheck.Output()
	if err != nil {
		fmt.Println("Erreur :", err)
		return
	}

	// Supprimer les lignes vides
	lines := []string{}
	for _, l := range strings.Split(string(out), "\n") { // on split la sortie en lignes
		lTrim := strings.TrimSpace(l)
		if lTrim != "" {
			lines = append(lines, lTrim) // on ajoute les lignes non vides à notre tableau lines
		}
	}

	// Vérifier s’il y a un processus
	if len(lines) == 0 {
		fmt.Println("PID introuvable.")
		return
	}

	// Afficher info du processus
	fmt.Println("Processus trouvé :", lines[0])

	// Confirmation kill
	fmt.Print("Confirmer kill (yes/no) : ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)
	if strings.ToLower(confirm) != "yes" {
		fmt.Println("Abandon.")
		return
	}

	// Exécuter le kill sur le PID spécifié de manière sécurisée
	var cmdKill *exec.Cmd
	if osName == "windows" {
		cmdKill = exec.Command("taskkill", "/PID", pidStr, "/F", "/T") // /F force le kill, /T tue tout les processus enfants
	} else { // Sur MacOS/Linux, on utilise kill -9 pour forcer la terminaison
		cmdKill = exec.Command("kill", "-9", pidStr)
	}

	// Exécuter la commande de kill et vérifier les erreurs
	err = cmdKill.Run()
	if err != nil {
		fmt.Println("Erreur lors du kill :", err)
	} else {
		fmt.Println("Processus tué :", pidStr)
	}
}

// ------- 16 / 20 : SecureOps Menu Cross-Platform macOS (normalement) et Windows --------
// Choix E : SecureOps (verrouillage de fichiers, lecture seule, audit log)
func logAction(outDir, action string) {
	f, err := os.OpenFile(filepath.Join(outDir, "audit.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Erreur audit log:", err)
		return
	}
	defer f.Close()
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	f.WriteString(fmt.Sprintf("[%s] %s\n", timestamp, action))
}

// Cette fonction crée un fichier de lock pour verrouiller le fichier cible
func lockFile(outDir, filename string) error {
	lockPath := filepath.Join(outDir, filename+".lock")
	if _, err := os.Stat(lockPath); err == nil {
		return fmt.Errorf("fichier déjà verrouillé")
	}
	f, err := os.Create(lockPath)
	if err != nil {
		return err
	}
	defer f.Close()
	logAction(outDir, "LOCK "+filename)
	return nil
}

// Cette fonction supprime le fichier de lock pour déverrouiller le fichier cible
func unlockFile(outDir, filename string) error {
	lockPath := filepath.Join(outDir, filename+".lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		return fmt.Errorf("fichier non verrouillé")
	}
	err := os.Remove(lockPath)
	// Vérifier les erreurs de suppression du fichier de lock
	if err != nil {
		return err
	}
	logAction(outDir, "UNLOCK "+filename)
	return nil
}

// Cette fonction vérifie si le fichier est verrouillé en vérifiant l'existence du fichier en .lock
func isLocked(outDir, filename string) bool {
	lockPath := filepath.Join(outDir, filename+".lock")
	if _, err := os.Stat(lockPath); err == nil {
		return true
	}
	return false
}

// Mettre un fichier en lecture seule
func setReadOnly(path string) error {
	osName := runtime.GOOS
	switch osName {
	case "windows":
		p, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			return err
		} // On utilise SetFileAttributes pour ajouter l'attribut FILE_ATTRIBUTE_READONLY
		err = syscall.SetFileAttributes(p, syscall.FILE_ATTRIBUTE_READONLY)
		if err != nil {
			return err
		}
	default: // macOS / Linux
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		mode := info.Mode()
		newMode := mode &^ 0222 // retirer les droits écriture
		err = os.Chmod(path, newMode)
		if err != nil {
			return err
		}
	}
	fmt.Println("Fichier mis en lecture seule:", path)
	return nil
}

// Supprimer lecture seule cross-platform
func unsetReadOnly(path string) error {
	osName := runtime.GOOS
	switch osName {
	case "windows":
		p, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			return err
		}
		// On remet les attributs à normal pour supprimer le read-only
		err = syscall.SetFileAttributes(p, syscall.FILE_ATTRIBUTE_NORMAL)
		if err != nil {
			return err
		}
	default: // macOS / Linux
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		// On remet les droits d'écriture en utilisant un OR binaire avec 0222
		mode := info.Mode()
		newMode := mode | 0222 // remettre droits écriture
		err = os.Chmod(path, newMode)
		if err != nil {
			return err
		}
	}
	fmt.Println("Lecture seule supprimée:", path)
	return nil
}

// Vérifier permissions cross-platform
func checkPermissions(path string) {
	osName := runtime.GOOS
	switch osName {
	case "windows":
		p, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			fmt.Println("Erreur vérification permissions:", err)
			return
		}
		attrs, err := syscall.GetFileAttributes(p)
		// Vérifier les erreurs de GetFileAttributes
		if err != nil {
			fmt.Println("Erreur vérification permissions:", err)
			return
		}
		// Vérifier si le fichier a l'attribut lecture seule
		if attrs&syscall.FILE_ATTRIBUTE_READONLY != 0 {
			fmt.Println("WARN: fichier en lecture seule:", path)
		} else {
			fmt.Println("Fichier modifiable:", path)
		}
	default:
		info, err := os.Stat(path)
		// Vérifier les erreurs de Stat
		if err != nil {
			fmt.Println("Erreur vérification permissions:", err)
			return
		}
		// Vérifier les permissions d'écriture du fichier
		if info.Mode().Perm()&0222 == 0 {
			fmt.Println("WARN: fichier en lecture seule:", path)
		} else {
			fmt.Println("Fichier modifiable:", path)
		}
	}
}

// Menu pour le SecOps (verrouillage, lecture seule, audit log)
func secureOpsMenu(cfg Config, reader *bufio.Reader) {
	for {
		fmt.Println("\n--- SecureOps Menu ---")
		fmt.Println()
		fmt.Println("1) Verrouiller fichier")
		fmt.Println("2) Déverrouiller fichier")
		fmt.Println("3) Mettre en lecture seule")
		fmt.Println("4) Retirer lecture seule")
		fmt.Println("5) Vérifier permissions")
		fmt.Println("6) Retour menu principal")
		fmt.Println()
		fmt.Print("Choix: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		// Pour pouvoir quitter directement après avoir choisi 6
		if choice == "6" {
			return
		}

		// Sinon on demande le chemin juste après les choix 1 à 5
		fmt.Print("Chemin du fichier (laisser simple nom pour utiliser out/ par défaut) : ")
		inputPath, _ := reader.ReadString('\n')
		inputPath = strings.TrimSpace(inputPath)

		var fullPath string
		var name string

		// Si le chemin est absolu, on l'utilise tel quel
		if filepath.IsAbs(inputPath) {
			fullPath = inputPath
			name = filepath.Base(inputPath)
		} else { // Sinon, on vérifie d'abord dans le répertoire courant
			if _, err := os.Stat(inputPath); err == nil {
				fullPath = inputPath
				name = filepath.Base(inputPath)
			} else { // Si pas trouvé dans le répertoire courant, on regarde dans out/
				fullPath = filepath.Join(cfg.OutDir, inputPath)
				name = filepath.Base(fullPath)
			}
		}

		// Vérifier que le fichier existe avant de tenter les opérations
		if _, err := os.Stat(fullPath); err != nil {
			fmt.Println("Fichier introuvable :", fullPath)
			continue
		}

		// En fonction du choix, on appelle la fonction correspondante
		switch choice {
		case "1":
			if isLocked(cfg.OutDir, name) {
				fmt.Println("Fichier déjà verrouillé")
			} else if err := lockFile(cfg.OutDir, name); err != nil {
				fmt.Println("Erreur verrouillage:", err)
			} else {
				fmt.Println("Fichier verrouillé avec succès")
			}
		case "2":
			if err := unlockFile(cfg.OutDir, name); err != nil {
				fmt.Println("Erreur déverrouillage:", err)
			} else {
				fmt.Println("Fichier déverrouillé avec succès")
			}
		case "3":
			if err := setReadOnly(fullPath); err != nil {
				fmt.Println("Erreur:", err)
			}
		case "4":
			if err := unsetReadOnly(fullPath); err != nil {
				fmt.Println("Erreur:", err)
			}
		case "5":
			checkPermissions(fullPath)
		default:
			fmt.Println("Choix invalide")
		}
	}
}
