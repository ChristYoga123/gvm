// manager/manager.go
package manager

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/schollz/progressbar/v3"
)

// Definisikan sumber unduhan yang dibedakan berdasarkan OS.
var downloadSources = map[string]map[string]string{
	"go": {
		"linux":   "https://go.dev/dl/go%s.%s-%s.tar.gz",
		"darwin":  "https://go.dev/dl/go%s.%s-%s.tar.gz",
		"windows": "https://go.dev/dl/go%s.%s-%s.zip",
	},
	"python": {
		"windows": "https://www.python.org/ftp/python/%s/python-%s-embed-amd64.zip",
	},
}

// GetBaseDir mengembalikan path ke direktori home ~/.gvm
func GetBaseDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".gvm"), nil
}

// GetVersionsDir mengembalikan path ke direktori versions
func GetVersionsDir() (string, error) {
	base, err := GetBaseDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "versions"), nil
}

// ListVersions mencantumkan versi yang terinstal untuk suatu bahasa
func ListVersions(lang string) ([]string, error) {
	versionsDir, err := GetVersionsDir()
	if err != nil {
		return nil, err
	}

	langPath := filepath.Join(versionsDir, lang)
	entries, err := os.ReadDir(langPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var versions []string
	for _, entry := range entries {
		if entry.IsDir() {
			versions = append(versions, entry.Name())
		}
	}
	return versions, nil
}

// Install mengunduh dan mengekstrak versi bahasa pemrograman
func Install(lang, version string) error {
	versionsDir, err := GetVersionsDir()
	if err != nil {
		return fmt.Errorf("gagal mendapatkan direktori versi: %w", err)
	}

	installPath := filepath.Join(versionsDir, lang, version)
	if _, err := os.Stat(installPath); !os.IsNotExist(err) {
		fmt.Printf("Versi %s untuk %s sudah terinstal di %s\n", version, lang, installPath)
		return nil
	}

	osSources, langSupported := downloadSources[lang]
	if !langSupported {
		return fmt.Errorf("instalasi otomatis untuk '%s' belum didukung di contoh ini", lang)
	}

	urlTemplate, osSupported := osSources[runtime.GOOS]
	if !osSupported {
		return fmt.Errorf("sistem operasi '%s' tidak didukung untuk instalasi otomatis %s", runtime.GOOS, lang)
	}

	url := fmt.Sprintf(urlTemplate, version, version)
	if lang == "go" {
		url = fmt.Sprintf(urlTemplate, version, runtime.GOOS, runtime.GOARCH)
	}

	fmt.Printf("Mengunduh %s versi %s dari %s\n", lang, version, url)

	fileSuffix := ".tmp"
	if strings.HasSuffix(url, ".zip") {
		fileSuffix = "*.zip"
	} else if strings.HasSuffix(url, ".tar.gz") {
		fileSuffix = "*.tar.gz"
	}

	tmpFile, err := os.CreateTemp("", fileSuffix)
	if err != nil {
		return fmt.Errorf("gagal membuat file sementara: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("gagal mengunduh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gagal mengunduh: status code %d", resp.StatusCode)
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"mengunduh",
	)
	io.Copy(io.MultiWriter(tmpFile, bar), resp.Body)
	fmt.Println("\nUnduhan selesai. Mengekstrak...")

	if strings.HasSuffix(url, ".zip") {
		if err := Unzip(tmpFile.Name(), installPath); err != nil {
			return fmt.Errorf("gagal mengekstrak arsip .zip: %w", err)
		}
	} else {
		tmpFile.Seek(0, 0)
		if err := Untar(tmpFile, installPath); err != nil {
			return fmt.Errorf("gagal mengekstrak arsip .tar.gz: %w", err)
		}
	}

	fmt.Printf("✅ Berhasil menginstal %s versi %s\n", lang, version)
	return nil
}

// ListAvailableVersions mengambil daftar versi yang tersedia dari sumber online.
func ListAvailableVersions(lang string) ([]string, error) {
	switch lang {
	case "python":
		return listPythonVersions()
	default:
		return nil, fmt.Errorf("pencarian versi otomatis untuk '%s' tidak didukung", lang)
	}
}

func listPythonVersions() ([]string, error) {
	fmt.Println("Mengambil daftar versi dari https://www.python.org/ftp/python/...")
	resp, err := http.Get("https://www.python.org/ftp/python/")
	if err != nil {
		return nil, fmt.Errorf("gagal mengambil data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status tidak ok: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca body: %w", err)
	}

	re := regexp.MustCompile(`href="([0-9]+\.[0-9]+\.[0-9]+)/"`)
	matches := re.FindAllStringSubmatch(string(body), -1)

	if len(matches) == 0 {
		return nil, fmt.Errorf("tidak ada versi yang ditemukan, mungkin pola regex perlu diperbarui")
	}

	rawVersions := make([]*version.Version, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			v, err := version.NewVersion(match[1])
			if err == nil {
				rawVersions = append(rawVersions, v)
			}
		}
	}

	sort.Sort(sort.Reverse(version.Collection(rawVersions)))

	sortedVersions := make([]string, len(rawVersions))
	for i, v := range rawVersions {
		sortedVersions[i] = v.Original()
	}

	return sortedVersions, nil
}

// GenerateUseCommand menghasilkan perintah shell untuk mengubah PATH
func GenerateUseCommand(lang, version string) (string, error) {
	versionsDir, err := GetVersionsDir()
	if err != nil {
		return "", err
	}

	versionPath := filepath.Join(versionsDir, lang, version)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		return "", fmt.Errorf("versi %s untuk %s tidak terinstal. Jalankan 'gvm install %s %s'", version, lang, lang, version)
	}

	var binPath string
	switch lang {
	case "go":
		binPath = filepath.Join(versionPath, "go", "bin")
	case "python":
		binPath = versionPath
	default:
		return "", fmt.Errorf("bahasa '%s' tidak dikenali untuk perintah 'use'", lang)
	}

	originalPath := os.Getenv("PATH")
	pathParts := filepath.SplitList(originalPath)
	baseDir, _ := GetBaseDir()

	var newPathParts []string
	newPathParts = append(newPathParts, binPath)

	for _, part := range pathParts {
		if !strings.Contains(part, baseDir) && part != "" {
			newPathParts = append(newPathParts, part)
		}
	}

	finalPath := strings.Join(newPathParts, string(os.PathListSeparator))

	if runtime.GOOS == "windows" {
		// MENGHASILKAN PERINTAH POWERSHELL YANG BENAR
		return fmt.Sprintf("$env:PATH = \"%s\"", finalPath), nil
	}
	return fmt.Sprintf("export PATH=\"%s\"", finalPath), nil
}

// Untar mengekstrak arsip .tar.gz
func Untar(r io.Reader, dest string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
}

// Unzip mengekstrak arsip .zip
func Unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("ilegal path file: %s", fpath)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
