package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

// DatabaseCall représente un appel à une base de données détecté dans le code.
type DatabaseCall struct {
	Function    string
	Line        uint32
	Description string
}

// Detection regroupe les informations relatives à une vulnérabilité détectée.
type Detection struct {
	CVE     string
	Line    uint32
	Message string
}

// PHPAnalyzer encapsule le parseur et fournit des méthodes pour analyser le code PHP.
type PHPAnalyzer struct {
	parser *sitter.Parser
}

// NewPHPAnalyzer crée et initialise un analyseur pour le langage PHP.
func NewPHPAnalyzer() *PHPAnalyzer {
	p := sitter.NewParser()
	p.SetLanguage(php.GetLanguage())
	return &PHPAnalyzer{parser: p}
}

// ParseFile lit et parse un fichier PHP, renvoyant son AST et le contenu source.
func (pa *PHPAnalyzer) ParseFile(filePath string) (*sitter.Tree, []byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	tree, err := pa.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, content, err
	}
	return tree, content, nil
}

// traverseAST effectue un parcours récursif de l’AST en appliquant la fonction visit à chaque nœud.
func traverseAST(node *sitter.Node, visit func(node *sitter.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		traverseAST(child, visit)
	}
}

// CountBranches retourne le nombre de branchements (if, while, for, foreach) dans l’AST.
func (pa *PHPAnalyzer) CountBranches(root *sitter.Node) int {
	count := 0
	branchTypes := map[string]bool{
		"if_statement":      true,
		"while_statement":   true,
		"for_statement":     true,
		"foreach_statement": true,
	}
	traverseAST(root, func(n *sitter.Node) {
		if branchTypes[n.Type()] {
			count++
		}
	})
	return count
}

// DetectDatabaseCalls recherche dans l’AST les appels susceptibles d’interagir avec une base de données.
func (pa *PHPAnalyzer) DetectDatabaseCalls(root *sitter.Node, source []byte) []DatabaseCall {
	var calls []DatabaseCall
	traverseAST(root, func(n *sitter.Node) {
		if n.Type() == "function_call_expression" || n.Type() == "member_call_expression" {
			funcName := extractFunctionName(n, source)
			line := n.StartPoint().Row + 1
			switch funcName {
			case "mysql_query", "mysqli_query":
				calls = append(calls, DatabaseCall{
					Function:    funcName,
					Line:        line,
					Description: fmt.Sprintf("Appel trouvé : %s", funcName),
				})
			case "execute":
				if n.Parent() != nil && n.Parent().Type() == "member_call_expression" {
					calls = append(calls, DatabaseCall{
						Function:    "$object->execute()",
						Line:        line,
						Description: "Appel trouvé : $object->execute()",
					})
				}
			case "exec":
				codeSnippet := string(source[n.StartByte():n.EndByte()])
				if strings.Contains(codeSnippet, "->mysql->exec") {
					calls = append(calls, DatabaseCall{
						Function:    "$object->mysql->exec",
						Line:        line,
						Description: "Appel trouvé : $object->mysql->exec(*)",
					})
				}
			}
		}
	})
	return calls
}

// DetectVulnerabilities parcourt l’AST à la recherche de vulnérabilités connues (CVEs).
func (pa *PHPAnalyzer) DetectVulnerabilities(root *sitter.Node, source []byte) []Detection {
	var detections []Detection
	traverseAST(root, func(n *sitter.Node) {
		if n.Type() == "function_call_expression" || n.Type() == "member_call_expression" {
			funcName := extractFunctionName(n, source)
			line := n.StartPoint().Row + 1
			switch funcName {
			// CVE-2017-7189 : fsockopen avec port confusion (exemple sur UDP)
			case "fsockopen":
				if isFsockopenPortConfusion(n, source) {
					detections = append(detections, Detection{
						CVE:     "CVE-2017-7189",
						Line:    line,
						Message: "fsockopen UDP détecté avec conflit de port",
					})
				}
			// CVE-2019-9025 : mb_split avec "\w" en premier argument
			case "mb_split":
				if isMbSplitW(n, source) {
					detections = append(detections, Detection{
						CVE:     "CVE-2019-9025",
						Line:    line,
						Message: `mb_split("\w") détecté`,
					})
				}
			// CVE-2019-11039 : iconv_mime_decode_headers détecté
			case "iconv_mime_decode_headers":
				detections = append(detections, Detection{
					CVE:     "CVE-2019-11039",
					Line:    line,
					Message: "iconv_mime_decode_headers(...) détecté",
				})
			// CVE-2020-7069 : openssl_encrypt avec AES-GCM/CCM
			case "openssl_encrypt":
				if isUsingGCmorCCM(n, source) {
					detections = append(detections, Detection{
						CVE:     "CVE-2020-7069",
						Line:    line,
						Message: "openssl_encrypt avec AES-GCM/CCM détecté",
					})
				}
			// CVE-2020-7071 / CVE-2021-21705 : filter_var avec FILTER_VALIDATE_URL
			case "filter_var":
				if isFilterVarValidateURL(n, source) {
					detections = append(detections, Detection{
						CVE:     "CVE-2020-7071 / CVE-2021-21705",
						Line:    line,
						Message: "filter_var(..., FILTER_VALIDATE_URL) détecté",
					})
				}
			// CVE-2021-21707 : simplexml_load_file avec chemin dynamique
			case "simplexml_load_file":
				if isSimplexmlLoadDynamic(n, source) {
					detections = append(detections, Detection{
						CVE:     "CVE-2021-21707",
						Line:    line,
						Message: "simplexml_load_file avec chemin dynamique détecté",
					})
				}
			}
		}
	})
	return detections
}

// AnalyzeDirectory parcourt récursivement un dossier et analyse chaque fichier PHP pour détecter des vulnérabilités.
// Aucun message n'est affiché si aucun résultat n'est trouvé.
func (pa *PHPAnalyzer) AnalyzeDirectory(dirPath string) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Erreur d'accès à %q: %v", path, err)
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".php") {
			return nil
		}

		tree, content, err := pa.ParseFile(path)
		if err != nil {
			log.Printf("Erreur d'analyse du fichier %q: %v", path, err)
			return nil
		}

		detections := pa.DetectVulnerabilities(tree.RootNode(), content)
		if len(detections) > 0 {
			fmt.Printf("\nAnalyse du fichier : %s\n", path)
			for _, d := range detections {
				fmt.Printf("[%s] %s (ligne %d)\n", d.CVE, d.Message, d.Line)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Erreur lors de la traversée du dossier %q: %v", dirPath, err)
	}
}

// AnalyzeDirectoryDBCalls parcourt récursivement un dossier et analyse chaque fichier PHP
// pour détecter les appels à la base de données.
// Aucun message n'est affiché si aucun appel n'est trouvé.
func (pa *PHPAnalyzer) AnalyzeDirectoryDBCalls(dirPath string) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Erreur d'accès à %q: %v", path, err)
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(strings.ToLower(info.Name()), ".php") {
			return nil
		}

		tree, content, err := pa.ParseFile(path)
		if err != nil {
			log.Printf("Erreur d'analyse du fichier %q: %v", path, err)
			return nil
		}

		calls := pa.DetectDatabaseCalls(tree.RootNode(), content)
		if len(calls) > 0 {
			fmt.Printf("\nAnalyse du fichier : %s\n", path)
			for _, call := range calls {
				fmt.Printf("- %s (ligne %d)\n", call.Description, call.Line)
			}
		}
		return nil
	})
	if err != nil {
		log.Printf("Erreur lors de la traversée du dossier %q: %v", dirPath, err)
	}
}

// extractFunctionName retourne le nom de la fonction pour un nœud d'appel (function ou member).
func extractFunctionName(node *sitter.Node, source []byte) string {
	if node.Type() == "function_call_expression" {
		if node.ChildCount() > 0 {
			fnChild := node.Child(0)
			if fnChild.Type() == "name" || fnChild.Type() == "qualified_name" {
				return string(source[fnChild.StartByte():fnChild.EndByte()])
			}
		}
	} else if node.Type() == "member_call_expression" {
		if nameNode := node.ChildByFieldName("name"); nameNode != nil {
			return string(source[nameNode.StartByte():nameNode.EndByte()])
		}
	}
	return ""
}

// getArguments extrait la liste brute des arguments passés à une fonction.
func getArguments(node *sitter.Node, source []byte) []string {
	var args []string
	argsNode := node.ChildByFieldName("arguments")
	if argsNode == nil {
		return args
	}
	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		if child.IsNamed() {
			raw := string(source[child.StartByte():child.EndByte()])
			args = append(args, raw)
		}
	}
	return args
}

// isFsockopenPortConfusion vérifie si le premier argument est une URL UDP contenant déjà un port
// et si un second argument numérique (port) est fourni.
func isFsockopenPortConfusion(node *sitter.Node, source []byte) bool {
	args := getArguments(node, source)
	if len(args) < 2 {
		return false
	}
	hostArg := args[0]
	portArg := args[1]
	isUDP := strings.Contains(strings.ToLower(hostArg), "udp://") && strings.Contains(hostArg, ":")
	isPortNumeric, _ := regexp.MatchString(`^\d+$`, portArg)
	return isUDP && isPortNumeric
}

// isMbSplitW vérifie si le premier argument vaut littéralement "\w".
func isMbSplitW(node *sitter.Node, source []byte) bool {
	args := getArguments(node, source)
	if len(args) == 0 {
		return false
	}
	firstArg := args[0]
	return firstArg == `"\w"`
}

// isUsingGCmorCCM vérifie si openssl_encrypt utilise un cipher contenant "gcm" ou "ccm".
func isUsingGCmorCCM(node *sitter.Node, source []byte) bool {
	args := getArguments(node, source)
	if len(args) < 2 {
		return false
	}
	cipherArg := strings.Trim(strings.ToLower(args[1]), `"' `)
	return strings.Contains(cipherArg, "-gcm") || strings.Contains(cipherArg, "-ccm")
}

// isFilterVarValidateURL vérifie que le deuxième argument de filter_var correspond à FILTER_VALIDATE_URL.
func isFilterVarValidateURL(node *sitter.Node, source []byte) bool {
	args := getArguments(node, source)
	if len(args) < 2 {
		return false
	}
	secondArg := args[1]
	return strings.Contains(secondArg, "FILTER_VALIDATE_URL")
}

// isSimplexmlLoadDynamic vérifie si le premier argument de simplexml_load_file est une variable (chemin dynamique).
func isSimplexmlLoadDynamic(node *sitter.Node, source []byte) bool {
	args := getArguments(node, source)
	if len(args) == 0 {
		return false
	}
	firstArg := args[0]
	return strings.HasPrefix(firstArg, "$")
}

func printUsage() {
	usage := `Usage: php-analyzer <command> [options]

Commands:
  count       - Compte les branchements dans un fichier PHP.
                Options:
                  -file string    Chemin vers le fichier PHP à analyser.

  dbcalls     - Détecte les appels à la base de données.
                Options:
                  -file string    Chemin vers le fichier PHP à analyser.
                  -dir  string    Chemin vers le dossier à analyser récursivement.

  cve         - Détecte les vulnérabilités (CVE) dans un fichier PHP.
                Options:
                  -file string    Chemin vers le fichier PHP à analyser.

  analyze-dir - Analyse récursivement un dossier contenant des fichiers PHP
                à la recherche de vulnérabilités.
                Options:
                  -dir string     Chemin vers le dossier à analyser.

Exemples:
  php-analyzer count -file=/chemin/vers/fichier.php
  php-analyzer dbcalls -file=/chemin/vers/fichier.php
  php-analyzer dbcalls -dir=/chemin/vers/dossier
  php-analyzer cve -file=/chemin/vers/fichier.php
  php-analyzer analyze-dir -dir=/chemin/vers/dossier
`
	fmt.Println(usage)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	analyzer := NewPHPAnalyzer()

	switch command {
	case "count":
		countCmd := flag.NewFlagSet("count", flag.ExitOnError)
		filePath := countCmd.String("file", "", "Chemin vers le fichier PHP à analyser")
		countCmd.Parse(os.Args[2:])
		if *filePath == "" {
			fmt.Println("Le flag -file est requis pour la commande count.")
			countCmd.Usage()
			os.Exit(1)
		}
		tree, _, err := analyzer.ParseFile(*filePath)
		if err != nil {
			log.Fatalf("Erreur lors du parsing du fichier %q: %v", *filePath, err)
		}
		branches := analyzer.CountBranches(tree.RootNode())
		if branches > 0 {
			fmt.Printf("Nombre de branchements dans %q : %d\n", *filePath, branches)
		}

	case "dbcalls":
		dbCmd := flag.NewFlagSet("dbcalls", flag.ExitOnError)
		filePath := dbCmd.String("file", "", "Chemin vers le fichier PHP à analyser")
		dirPath := dbCmd.String("dir", "", "Chemin vers le dossier à analyser récursivement")
		dbCmd.Parse(os.Args[2:])

		if *filePath == "" && *dirPath == "" {
			fmt.Println("Le flag -file ou -dir est requis pour la commande dbcalls.")
			dbCmd.Usage()
			os.Exit(1)
		}

		// Analyse d'un fichier
		if *filePath != "" {
			tree, content, err := analyzer.ParseFile(*filePath)
			if err != nil {
				log.Fatalf("Erreur lors du parsing du fichier %q: %v", *filePath, err)
			}
			calls := analyzer.DetectDatabaseCalls(tree.RootNode(), content)
			if len(calls) > 0 {
				fmt.Printf("Appels de base de données détectés dans %q :\n", *filePath)
				for _, call := range calls {
					fmt.Printf("- %s (ligne %d)\n", call.Description, call.Line)
				}
			}
		}

		// Analyse d'un dossier récursif
		if *dirPath != "" {
			analyzer.AnalyzeDirectoryDBCalls(*dirPath)
		}

	case "cve":
		cveCmd := flag.NewFlagSet("cve", flag.ExitOnError)
		filePath := cveCmd.String("file", "", "Chemin vers le fichier PHP à analyser")
		cveCmd.Parse(os.Args[2:])
		if *filePath == "" {
			fmt.Println("Le flag -file est requis pour la commande cve.")
			cveCmd.Usage()
			os.Exit(1)
		}
		tree, content, err := analyzer.ParseFile(*filePath)
		if err != nil {
			log.Fatalf("Erreur lors du parsing du fichier %q: %v", *filePath, err)
		}
		detections := analyzer.DetectVulnerabilities(tree.RootNode(), content)
		if len(detections) > 0 {
			for _, d := range detections {
				fmt.Printf("[%s] %s (ligne %d)\n", d.CVE, d.Message, d.Line)
			}
		}

	case "analyze-dir":
		dirCmd := flag.NewFlagSet("analyze-dir", flag.ExitOnError)
		dirPath := dirCmd.String("dir", "", "Chemin vers le dossier à analyser")
		dirCmd.Parse(os.Args[2:])
		if *dirPath == "" {
			fmt.Println("Le flag -dir est requis pour la commande analyze-dir.")
			dirCmd.Usage()
			os.Exit(1)
		}
		analyzer.AnalyzeDirectory(*dirPath)

	default:
		fmt.Printf("Commande inconnue : %q\n", command)
		printUsage()
		os.Exit(1)
	}
}
