// main.go
package main

import (
	"fmt"
	"gvm-project/manager"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gvm",
		Short: "gvm (Generic Version Manager) adalah manajer versi multi-bahasa.",
		Long:  `Sebuah alat CLI yang terinspirasi oleh nvm untuk mengelola beberapa versi dari berbagai bahasa pemrograman.`,
	}

	// Perintah: gvm install <bahasa> <versi>
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

	// Perintah: gvm list <bahasa>
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

	// Perintah: gvm use <bahasa> <versi>
	var useCmd = &cobra.Command{
		Use:   "use [bahasa] [versi]",
		Short: "Gunakan versi bahasa tertentu di shell saat ini",
		Long:  `Penting: Gunakan perintah ini dengan eval, contoh: eval "$(gvm use go 1.22.3)"`,
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			lang := args[0]
			version := args[1]
			shellCommand, err := manager.GenerateUseCommand(lang, version)
			if err != nil {
				// Cetak error ke stderr agar tidak ikut dievaluasi oleh eval
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Cetak perintah ke stdout agar bisa dieksekusi oleh eval
			fmt.Println(shellCommand)
		},
	}

	rootCmd.AddCommand(installCmd, listCmd, useCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
