// manager/manager.go
package manager

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/schollz/progressbar/v3"
)

// Definisikan sumber unduhan. Dalam proyek nyata, ini bisa dari file konfigurasi JSON/YAML.
var downloadSources = map[string]string{
	"go":     "https://go.dev/dl/go%s.%s-%s.tar.gz",                   // versi, os, arch
	"python": "https://www.python.org/ftp/python/%s/Python-%s.tar.xz", // Ini lebih rumit, seringkali perlu build dari source. Contoh ini menyederhanakan.
	// Untuk Python, lebih mudah menggunakan pre-built binaries, tapi URL-nya sangat bervariasi.
	// Untuk PHP, juga bervariasi.
	// Kita akan fokus pada Go untuk contoh unduhan yang berfungsi penuh.
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
		// Jika direktori tidak ada, berarti belum ada versi yang terinstal
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

	// Contoh ini hanya mengimplementasikan untuk Go karena URL-nya konsisten.
	if lang != "go" {
		return fmt.Errorf("instalasi otomatis untuk '%s' belum didukung di contoh ini", lang)
	}

	url := fmt.Sprintf(downloadSources[lang], version, runtime.GOOS, runtime.GOARCH)
	fmt.Printf("Mengunduh %s versi %s dari %s\n", lang, version, url)

	// Buat file sementara
	tmpFile, err := os.CreateTemp("", "*.tar.gz")
	if err != nil {
		return fmt.Errorf("gagal membuat file sementara: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	// Unduh file dengan progress bar
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("gagal mengunduh: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("gagal mengunduh: status code %d", resp.StatusCode)
	}

	f, _ := os.OpenFile(tmpFile.Name(), os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"mengunduh",
	)
	io.Copy(io.MultiWriter(f, bar), resp.Body)

	fmt.Println("\nUnduhan selesai. Mengekstrak...")

	// Kembali ke awal file untuk dibaca
	tmpFile.Seek(0, 0)

	// Ekstrak arsip
	if err := Untar(tmpFile, installPath); err != nil {
		return fmt.Errorf("gagal mengekstrak arsip: %w", err)
	}

	fmt.Printf("✅ Berhasil menginstal %s versi %s\n", lang, version)
	return nil
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

	// Path ke direktori bin bervariasi antar bahasa
	var binPath string
	switch lang {
	case "go":
		// Instalasi go dari tar.gz akan membuat folder 'go' di dalamnya
		binPath = filepath.Join(versionPath, "go", "bin")
	case "python":
		binPath = filepath.Join(versionPath, "bin")
	// Tambahkan case lain untuk php, java, dll.
	default:
		return "", fmt.Errorf("bahasa '%s' tidak dikenali untuk perintah 'use'", lang)
	}

	// Membersihkan PATH dari instalasi gvm lainnya
	originalPath := os.Getenv("PATH")
	pathParts := filepath.SplitList(originalPath)

	var newPathParts []string
	baseDir, _ := GetBaseDir()

	newPathParts = append(newPathParts, binPath) // Tambahkan path baru di depan

	for _, part := range pathParts {
		// Jika path lama bukan bagian dari gvm, pertahankan
		if !strings.Contains(part, baseDir) {
			newPathParts = append(newPathParts, part)
		}
	}

	finalPath := strings.Join(newPathParts, string(os.PathListSeparator))

	// Mencetak perintah untuk dievaluasi oleh shell
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

	// Pastikan direktori tujuan ada
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
			// Pastikan direktori untuk file ini ada
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
