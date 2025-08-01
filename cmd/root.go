package cmd

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"bbs/internal/config"
	"bbs/internal/database"
	"bbs/internal/server"
	"bbs/internal/terminal"
)

var (
	cfgFile   string
	localMode bool
)

var rootCmd = &cobra.Command{
	Use:   "coastline-bbs",
	Short: "A classic BBS experience over SSH",
	Long: `Coastline BBS is a modern recreation of classic bulletin board 
systems, accessible via SSH with authentic terminal-based interaction.

Run without flags to start the SSH server, or use -l/--local to 
connect directly to the BBS in your current terminal.`,
	Run: func(cmd *cobra.Command, args []string) {
		if localMode {
			runLocalMode()
		} else {
			runServerMode()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	rootCmd.Flags().BoolVarP(&localMode, "local", "l", false, "Run in local terminal mode instead of starting SSH server")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

func runLocalMode() {
	configFile := "config.yaml"
	if cfgFile != "" {
		configFile = cfgFile
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.Initialize(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	term := terminal.NewLocalTerminal()
	defer term.Close()

	fmt.Println("Starting local BBS session...")
	fmt.Println("Press Ctrl+C to exit")

	// Use unified server
	bbsServer := server.NewServer(cfg, db)
	session := bbsServer.NewLocalSession(term)
	session.Run()
}

func runServerMode() {
	configFile := "config.yaml"
	if cfgFile != "" {
		configFile = cfgFile
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.Initialize(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Use unified server for SSH
	bbsServer := server.NewServer(cfg, db)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", cfg.Server.Port, err)
	}
	defer listener.Close()

	log.Printf("Coastline BBS Server listening on port %d", cfg.Server.Port)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
			go bbsServer.HandleConnection(conn)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down server...")
}
