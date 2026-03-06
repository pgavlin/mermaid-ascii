package cmd

import (
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/pgavlin/mermaid-ascii/pkg/diagram"
	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

// Add the resultCache variable with additional fields
var (
	resultCache = struct {
		sync.RWMutex
		m map[string]cacheEntry
	}{m: make(map[string]cacheEntry)}
	maxCacheSize = 10000 // Maximum number of entries in the cache
)

type cacheEntry struct {
	value string
}

var (
	gitVersion     string
	gitVersionOnce sync.Once
)

func init() {
	rootCmd.AddCommand(webCmd)
}

var webCmd = &cobra.Command{
	Use:   "web",
	Short: "HTTP server for rendering mermaid diagrams.",
	Run: func(cmd *cobra.Command, args []string) {
		if Verbose {
			log.SetLevel(log.DebugLevel)
		} else {
			log.SetLevel(log.InfoLevel)
		}
		r := setupRouter()
		// Listen and Server in 0.0.0.0:8080
		err := r.Run(":3001")
		if err != nil {
			panic(err)
		}
	},
}

// Add this function near the top of the file, after the imports
func getGitVersion() string {
	gitVersionOnce.Do(func() {
		log.Info("Getting git version")
		cmd := exec.Command("git", "describe", "--tags", "--always")
		output, err := cmd.Output()
		if err != nil {
			log.Warnf("Failed to get git version: %v", err)
			gitVersion = "unknown"
		} else {
			gitVersion = strings.TrimSpace(string(output))
		}
	})
	return gitVersion
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"Version": getGitVersion(),
		})
	})

	r.POST("/", renderMermaid)

	// Backwards compatibility
	r.POST("/generate", renderMermaid)

	return r
}

func renderMermaid(c *gin.Context) {
	mermaidString := c.PostForm("mermaid")
	// Parse xPadding and yPadding as integers
	xPaddingStr := c.PostForm("xPadding")
	pX := cliPaddingBetweenX
	if xPaddingStr != "" {
		if padding, err := strconv.Atoi(xPaddingStr); err == nil {
			pX = padding
		} else {
			log.Warnf("Invalid xPadding value: %s", xPaddingStr)
		}
	}

	yPaddingStr := c.PostForm("yPadding")
	pY := cliPaddingBetweenY
	if yPaddingStr != "" {
		if padding, err := strconv.Atoi(yPaddingStr); err == nil {
			pY = padding
		} else {
			log.Warnf("Invalid yPadding value: %s", yPaddingStr)
		}
	}
	useExtendedCharsData := c.PostForm("useExtendedChars")
	useAsciiMode := useExtendedCharsData == ""
	log.Debugf("Received input %s", c.Request.PostForm.Encode())

	// Create a cache key using the input parameters
	cacheKey := mermaidString + "x" + xPaddingStr + "y" + yPaddingStr + "e" + useExtendedCharsData

	// Check if the result is already in the cache
	resultCache.RLock()
	entry, found := resultCache.m[cacheKey]
	resultCache.RUnlock()

	if found {
		log.Infof("Cache hit for key: %s", cacheKey)
		c.String(http.StatusOK, entry.value)
		return
	}

	// Create render configuration
	config, err := diagram.NewWebConfig(
		useAsciiMode,
		cliBoxBorderPadding,
		pX,
		pY,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid configuration: %v", err)})
		return
	}
	config.Verbose = Verbose // Allow verbose logging in web mode if enabled
	result, err := RenderDiagram(mermaidString, config)
	if err != nil {
		log.Errorf("Rendering failed: %v", err)
		c.String(http.StatusBadRequest, fmt.Sprintf("Failed to render diagram: %v", err))
		return
	}

	// Store the result in the cache
	resultCache.Lock()
	if len(resultCache.m) >= maxCacheSize {
		log.Infof("Cache is full, removing oldest entry")
		// Remove a random entry if cache is full
		for k := range resultCache.m {
			delete(resultCache.m, k)
			break
		}
	}
	resultCache.m[cacheKey] = cacheEntry{
		value: result,
	}
	resultCache.Unlock()

	c.String(http.StatusOK, result)
}
