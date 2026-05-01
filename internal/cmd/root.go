package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/killakam3084/rarclean/internal/config"
	"github.com/killakam3084/rarclean/internal/extractor"
	"github.com/killakam3084/rarclean/internal/mover"
	"github.com/killakam3084/rarclean/internal/qbittorrent"
	"github.com/spf13/cobra"
)

var (
	configPath string
	dryRun     bool
	targetPath string
)

var rootCmd = &cobra.Command{
	Use:   "rarclean",
	Short: "Automated RAR extraction and qBittorrent management for media libraries",
	Long: `rarclean is a CLI tool that automates the workflow of:
1. Locating RAR archives in a directory
2. Extracting them using 7z
3. Finding the associated torrent in qBittorrent
4. Moving the RAR directory to a zombies location
5. Updating qBittorrent's file tracking

This maintains seeding while organizing extracted media.`,
	RunE: runRarclean,
}

func init() {
	rootCmd.Flags().StringVar(&configPath, "config", "config.json", "Path to configuration file")
	rootCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show planned actions without executing")
	rootCmd.Flags().StringVar(&targetPath, "path", "", "Directory containing RAR files to process (required)")
	rootCmd.MarkFlagRequired("path")
}

func Execute() error {
	return rootCmd.Execute()
}

func runRarclean(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	fmt.Println("=== rarclean - RAR Extraction & qBittorrent Manager ===\n")

	// Step 1: Find RAR files
	fmt.Printf("Step 1: Scanning directory for RAR files: %s\n", targetPath)
	ext := extractor.New(targetPath)
	rarFiles, err := ext.FindRARFiles()
	if err != nil {
		return fmt.Errorf("failed to find RAR files: %w", err)
	}
	if len(rarFiles) == 0 {
		fmt.Println("No RAR files found in directory")
		return nil
	}

	fmt.Printf("Found %d RAR archive(s):\n", len(rarFiles))
	for i, rar := range rarFiles {
		fmt.Printf("  %d. %s (in %s)\n", i+1, filepath.Base(rar.Path), filepath.Base(rar.Directory))
	}
	fmt.Println()

	// Initialize qBittorrent client
	qbClient, err := qbittorrent.New(cfg.QBittorrent.URL, cfg.QBittorrent.Username, cfg.QBittorrent.Password)
	if err != nil {
		return fmt.Errorf("failed to initialize qBittorrent client: %w", err)
	}

	// Login to qBittorrent
	fmt.Println("Authenticating with qBittorrent...")
	if err := qbClient.Login(); err != nil {
		return fmt.Errorf("failed to authenticate with qBittorrent: %w", err)
	}
	fmt.Println()

	// Process each RAR file
	moveMgr := mover.New(dryRun)
	for i, rarFile := range rarFiles {
		fmt.Printf("Processing RAR archive %d of %d: %s\n", i+1, len(rarFiles), rarFile.BaseName)

		// Step 2: Extract
		fmt.Printf("  Step 2: Extracting with 7z...\n")
		if dryRun {
			fmt.Printf("    [DRY RUN] Would extract: %s\n", rarFile.Path)
		} else {
			if err := ext.Extract(rarFile); err != nil {
				fmt.Printf("    ERROR: Extraction failed: %v\n", err)
				continue
			}
		}
		fmt.Println()

		// Step 3: Find torrent
		fmt.Printf("  Step 3: Finding torrent in qBittorrent...\n")
		torrent, err := qbClient.FindTorrentByPath(rarFile.Directory)
		if err != nil {
			fmt.Printf("    ERROR: Failed to query torrents: %v\n", err)
			continue
		}
		if torrent == nil {
			fmt.Printf("    WARNING: No torrent found for path: %s\n", rarFile.Directory)
			fmt.Println("    (RAR files will remain in place)")
			fmt.Println()
			continue
		}
		fmt.Printf("    Found torrent: %s\n", torrent.Name)
		fmt.Printf("    Current location: %s\n", torrent.SavePath)
		fmt.Println()

		// Step 4: Relocate RAR directory
		fmt.Printf("  Step 4: Relocating RAR directory to zombies...\n")
		newLocation := filepath.Join(cfg.Paths.Zombies, filepath.Base(rarFile.Directory))
		fmt.Printf("    From: %s\n", rarFile.Directory)
		fmt.Printf("    To:   %s\n", newLocation)
		if err := moveMgr.ValidateMoveOperation(rarFile.Directory, newLocation); err != nil {
			fmt.Printf("    ERROR: Move validation failed: %v\n", err)
			continue
		}
		if err := moveMgr.MoveDirectory(rarFile.Directory, newLocation); err != nil {
			fmt.Printf("    ERROR: Move failed: %v\n", err)
			continue
		}
		fmt.Println()

		// Step 5: Update torrent location in qBittorrent
		fmt.Printf("  Step 5: Updating qBittorrent torrent location...\n")
		if err := qbClient.SetLocation(torrent.Hash, newLocation); err != nil {
			fmt.Printf("    ERROR: Failed to update torrent location: %v\n", err)
			fmt.Printf("    NOTE: RAR directory moved but torrent tracking may be broken\n")
			continue
		}

		// Verify the update
		updatedTorrent, err := qbClient.GetTorrent(torrent.Hash)
		if err != nil {
			fmt.Printf("    ERROR: Failed to verify update: %v\n", err)
			continue
		}
		if updatedTorrent != nil {
			fmt.Printf("    Updated torrent location: %s\n", updatedTorrent.SavePath)
		}

		fmt.Println("  ✓ Archive processed successfully")
		fmt.Println()
	}

	fmt.Println("=== Processing complete ===")
	if dryRun {
		fmt.Println("(This was a dry-run; no actual changes were made)")
	}

	return nil
}
