# PHP Analyzer Labo 1 LOG6302A

## Compilation

Il faut installer Go v1.22+ et tree-sitter

Pour compiler l'outil, exécutez :

```bash
go build -o php-analyzer main.go
```

Pour Windows, privilégiez l'utilisation du fichier précompilé `php-analyzer.exe` afin d'éviter d'éventuels problèmes de compatibilité liés à la compilation sur windows.

## Utilisation

L'outil se lance via la ligne de commande avec plusieurs sous-commandes. Voici les principales :

### 1. Compter les branchements

Commande : count

Description : Compte les structures de contrôle (if, while, for, foreach) dans un fichier PHP.

Exemple :

```bash
./php-analyzer count -file code_to_analyze/binary-search.php
```

```bash
Nombre de branchements dans "code_to_analyze/binary-search.php" : 4 
```

### 2. Détecter les appels à la base de données

Commande : dbcalls

Description : Détecte les appels aux fonctions de base de données. Vous pouvez analyser :

    Un seul fichier (avec -file)
    Un dossier récursivement (avec -dir)

Exemples :

```bash
./php-analyzer dbcalls -dir code_to_analyze/wordpress_sources/

Analyse du fichier : code_to_analyze/wordpress_sources/wp-includes/SimplePie/Cache/MySQL.php
- Appel trouvé : $object->mysql->exec(*) (ligne 130)
- Appel trouvé : $object->mysql->exec(*) (ligne 139)

Analyse du fichier : code_to_analyze/wordpress_sources/wp-includes/wp-db.php
- Appel trouvé : mysqli_query (ligne 830)
- Appel trouvé : mysql_query (ligne 840)
- Appel trouvé : mysqli_query (ligne 859)
- Appel trouvé : mysql_query (ligne 861)
- Appel trouvé : mysqli_query (ligne 905)
- Appel trouvé : mysql_query (ligne 907)
- Appel trouvé : mysqli_query (ligne 1877)
- Appel trouvé : mysql_query (ligne 1879)
```

### 3. Détecter des vulnérabilités

Commande : analyze-dir

Description : Analyse un fichier PHP pour détecter des vulnérabilités connues.

Exemple :

```bash
./php-analyzer analyze-dir -dir code_to_analyze/test_cve/

Analyse du fichier : code_to_analyze/test_cve/2017_7189.php
[CVE-2017-7189] fsockopen UDP détecté avec conflit de port (ligne 9)

Analyse du fichier : code_to_analyze/test_cve/2019_11039.php
[CVE-2019-11039] iconv_mime_decode_headers(...) détecté (ligne 21)

Analyse du fichier : code_to_analyze/test_cve/2019_9025.php
[CVE-2019-9025] mb_split("\w") détecté (ligne 8)

Analyse du fichier : code_to_analyze/test_cve/2020_7069.php
[CVE-2020-7069] openssl_encrypt avec AES-GCM/CCM détecté (ligne 11)

Analyse du fichier : code_to_analyze/test_cve/2020_7071.php
[CVE-2020-7071 / CVE-2021-21705] filter_var(..., FILTER_VALIDATE_URL) détecté (ligne 9)

Analyse du fichier : code_to_analyze/test_cve/2021_21705.php
[CVE-2020-7071 / CVE-2021-21705] filter_var(..., FILTER_VALIDATE_URL) détecté (ligne 9)

Analyse du fichier : code_to_analyze/test_cve/2021_21707.php
[CVE-2021-21707] simplexml_load_file avec chemin dynamique détecté (ligne 10)
```
