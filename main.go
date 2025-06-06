// main.go
package main

import (
	"fmt"
	"gvm-project/manager"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gvm",
		Short: "gvm (Generic Version Manager) adalah manajer versi multi-bahasa.",
		Long:  `Sebuah alat CLI yang terinspirasi oleh nvm untuk mengelola beberapa versi dari berbagai bahasa pemrograman.`,
	}

	var installCmd = &cobra.Command{
		Use:   "install [bahasa] [versi]",
		Short: "Install versi bahasa pemrograman tertentu",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			lang := args[0]
			version := args[1]
			if err := manager.Install(lang, version); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list [bahasa]",
		Short: "Tampilkan versi yang terinstal untuk suatu bahasa",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			lang := args[0]
			versions, err := manager.ListVersions(lang)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			if len(versions) == 0 {
				fmt.Printf("Belum ada versi %s yang terinstal.\n", lang)
				return
			}
			fmt.Printf("Versi %s yang terinstal:\n", lang)
			for _, v := range versions {
				fmt.Printf("- %s\n", v)
			}
		},
	}

	var searchCmd = &cobra.Command{
		Use:   "search [bahasa]",
		Short: "Cari versi online yang tersedia untuk suatu bahasa",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			lang := args[0]
			versions, err := manager.ListAvailableVersions(lang)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Versi %s yang tersedia (dari yang terbaru):\n", lang)
			for _, v := range versions {
				fmt.Println(v)
			}
		},
	}

	var useLongHelp string
	if runtime.GOOS == "windows" {
		useLongHelp = `Penting: Perintah ini mencetak perintah 'set'. Anda harus menjalankannya agar berpengaruh.
Contoh untuk CMD:
> for /f "tokens=*" %i in ('gvm use go 1.22.3') do %i

Contoh untuk PowerShell:
> gvm use go 1.22.3 | iex`
	} else {
		useLongHelp = `Penting: Gunakan perintah ini dengan eval agar berpengaruh pada shell saat ini.
Contoh: eval "$(gvm use go 1.22.3)"`
	}

	var useCmd = &cobra.Command{
		Use:   "use [bahasa] [versi]",
		Short: "Gunakan versi bahasa tertentu di shell saat ini",
		Long:  useLongHelp,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			lang := args[0]
			version := args[1]
			shellCommand, err := manager.GenerateUseCommand(lang, version)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(shellCommand)
		},
	}

	rootCmd.AddCommand(installCmd, listCmd, searchCmd, useCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
