package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"restaurant-backend/internal/branch"
)

const (
	defaultPort = "9002"
)

// ============================================================
// Context keys — نفس اللي عند Attendance API
// ============================================================
const (
	CtxUserID   = "user_id"
	CtxUserRole = "user_role"
	CtxBranchID = "branch_id"
)

func main() {
	// ============================================================
	// 1. Load .env
	// ============================================================
	if err := godotenv.Load(); err != nil {
		log.Printf("⚠️  .env not found, using environment variables")
	}

	// ============================================================
	// 2. Database
	// ============================================================
	dbURL := buildDBURL()
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatalf("cannot create db pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("cannot connect to database: %v", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// ============================================================
	// 3. Branch Module
	// ============================================================
	branchRepo := branch.NewPostgresRepository(pool)
	branchSvc := branch.NewService(branchRepo, branch.SystemClock{})
	branchHandler := branch.NewHandler(branchSvc)
	log.Println("✅ Branch module ready")

	// ============================================================
	// 4. Gin Router
	// ============================================================
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(apiGatewayMiddleware())

	// Health (public)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "branch-api",
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// API v1 (protected)
	api := router.Group("/api/v1")
	{
		branchHandler.RegisterRoutes(api)
	}

	// ============================================================
	// 5. Start Server
	// ============================================================
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("🚀 Branch API listening on :%s", port)
	log.Printf("   ❤️  Health:   http://localhost:%s/health", port)
	log.Printf("   🔗 BaseURL:  http://localhost:%s/api/v1", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

// ============================================================
// buildDBURL — يبني URL من .env
// ============================================================
func buildDBURL() string {
	// الأول: DATABASE_URL (لو موجود)
	if u := os.Getenv("DATABASE_URL"); u != "" {
		return u
	}

	// التاني: تركيب من الأجزاء
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || pass == "" || name == "" {
		log.Fatal("Missing database settings: set DATABASE_URL or DB_HOST/DB_PORT/DB_USER/DB_PASSWORD/DB_NAME")
	}

	connectionURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, pass),
		Host:   fmt.Sprintf("%s:%s", host, port),
		Path:   name,
	}
	q := connectionURL.Query()
	q.Set("sslmode", "disable")
	connectionURL.RawQuery = q.Encode()

	return connectionURL.String()
}

// ============================================================
// apiGatewayMiddleware — reads Gateway headers
// ============================================================
func apiGatewayMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if isPublicPath(path) {
			c.Next()
			return
		}

		userID := strings.TrimSpace(c.GetHeader("X-User-ID"))
		userRole := strings.TrimSpace(c.GetHeader("X-User-Role"))
		branchID := strings.TrimSpace(c.GetHeader("X-Branch-ID"))

		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "missing X-User-ID header",
				},
			})
			return
		}

		if userRole == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "missing X-User-Role header",
				},
			})
			return
		}

		c.Set(CtxUserID, userID)
		c.Set(CtxUserRole, userRole)
		c.Set(CtxBranchID, branchID)

		c.Next()
	}
}

func isPublicPath(path string) bool {
	return path == "/health" || strings.HasPrefix(path, "/swagger/")
}
