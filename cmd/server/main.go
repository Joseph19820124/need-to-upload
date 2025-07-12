package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	httpserver "github.com/github-mcp-http/internal/transport/http"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "github-mcp-http",
		Short: "GitHub MCP server with HTTP/SSE transport",
		Long:  "A Model Context Protocol server for GitHub operations using HTTP and Server-Sent Events",
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.github-mcp-http.yaml)")
	
	rootCmd.AddCommand(httpCmd)
}

var httpCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP/SSE server",
	Run:   runHTTPServer,
}

func init() {
	httpCmd.Flags().String("host", "0.0.0.0", "Host to bind to")
	httpCmd.Flags().Int("port", 8080, "Port to listen on")
	httpCmd.Flags().String("github-token", "", "GitHub Personal Access Token")
	httpCmd.Flags().Bool("read-only", false, "Enable read-only mode")
	httpCmd.Flags().String("tls-cert", "", "Path to TLS certificate")
	httpCmd.Flags().String("tls-key", "", "Path to TLS key")
	
	viper.BindPFlag("host", httpCmd.Flags().Lookup("host"))
	viper.BindPFlag("port", httpCmd.Flags().Lookup("port"))
	viper.BindPFlag("github.token", httpCmd.Flags().Lookup("github-token"))
	viper.BindPFlag("github.read_only", httpCmd.Flags().Lookup("read-only"))
	viper.BindPFlag("tls.cert", httpCmd.Flags().Lookup("tls-cert"))
	viper.BindPFlag("tls.key", httpCmd.Flags().Lookup("tls-key"))
}

func runHTTPServer(cmd *cobra.Command, args []string) {
	// Get GitHub token from environment variable first, fallback to viper
	githubToken := os.Getenv("GITHUB_MCP_GITHUB_TOKEN")
	if githubToken == "" {
		githubToken = viper.GetString("github.token")
	}

	// Get other config from environment variables with fallback to viper
	host := os.Getenv("GITHUB_MCP_HOST")
	if host == "" {
		host = viper.GetString("host")
	}

	port := viper.GetInt("port")
	// Railway provides PORT environment variable
	if railwayPort := os.Getenv("PORT"); railwayPort != "" {
		if p, err := strconv.Atoi(railwayPort); err == nil {
			port = p
		}
	} else if envPort := os.Getenv("GITHUB_MCP_PORT"); envPort != "" {
		if p, err := strconv.Atoi(envPort); err == nil {
			port = p
		}
	}

	readOnly := viper.GetBool("github.read_only")
	if envReadOnly := os.Getenv("GITHUB_MCP_GITHUB_READ_ONLY"); envReadOnly != "" {
		readOnly = envReadOnly == "true"
	}

	config := &httpserver.ServerConfig{
		Host:        host,
		Port:        port,
		TLSCert:     viper.GetString("tls.cert"),
		TLSKey:      viper.GetString("tls.key"),
		GitHubToken: githubToken,
		ReadOnly:    readOnly,
	}

	server, err := httpserver.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler:      server,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Starting HTTP/SSE server on %s", httpServer.Addr)
		var err error
		if config.TLSCert != "" && config.TLSKey != "" {
			err = httpServer.ListenAndServeTLS(config.TLSCert, config.TLSKey)
		} else {
			err = httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}
	log.Println("Server stopped")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".github-mcp-http")
	}

	viper.SetEnvPrefix("GITHUB_MCP")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}