package facade_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"

	appservice "github.com/angelapytao/diffgram-go/application/service"
	"github.com/angelapytao/diffgram-go/config"
	diffdb "github.com/angelapytao/diffgram-go/infrastructure/db"
	infratoken "github.com/angelapytao/diffgram-go/infrastructure/token"
	"github.com/angelapytao/diffgram-go/interfaces/http/facade"
	"github.com/angelapytao/diffgram-go/interfaces/http/middleware"
)

func startMySQLForE2E(t *testing.T) string {
	t.Helper()
	ctx := context.Background()
	container, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("diffgram_test"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("testpass"),
	)
	require.NoError(t, err)
	t.Cleanup(func() { _ = container.Terminate(ctx) })

	dsn, err := container.ConnectionString(ctx, "charset=utf8mb4&parseTime=True&loc=Local")
	require.NoError(t, err)
	return dsn
}

func TestE2E_RegisterLoginCreateProjectList(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test: requires Docker")
	}

	dsn := startMySQLForE2E(t)

	sqlDB, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	defer func() { _ = sqlDB.Close() }()

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations")
	require.NoError(t, diffdb.RunMigrations(sqlDB, migrationsDir))

	gormDB, err := diffdb.NewConnection(dsn)
	require.NoError(t, err)

	tokenSvc := infratoken.NewJWTService(config.JWTConfig{Secret: "e2e-secret", Timeout: time.Hour})
	appservice.Init(gormDB, tokenSvc, nil)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	facade.RegisterUserRoutes(r)
	authed := r.Group("/", middleware.Auth(tokenSvc))
	facade.RegisterProjectRoutes(authed)

	srv := httptest.NewServer(r)
	defer srv.Close()
	base := srv.URL

	client := &http.Client{Timeout: 10 * time.Second}

	// 1. Register
	regBody, _ := json.Marshal(map[string]string{
		"email":    "e2e@example.com",
		"password": "e2epassword1",
	})
	regResp, err := client.Post(base+"/api/v1/user/new", "application/json", bytes.NewReader(regBody))
	require.NoError(t, err)
	defer func() { _ = regResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, regResp.StatusCode, "register must succeed")

	// 2. Login
	loginBody, _ := json.Marshal(map[string]string{
		"email":    "e2e@example.com",
		"password": "e2epassword1",
		"mode":     "password",
	})
	loginResp, err := client.Post(base+"/api/user/login", "application/json", bytes.NewReader(loginBody))
	require.NoError(t, err)
	defer func() { _ = loginResp.Body.Close() }()
	require.Equal(t, http.StatusOK, loginResp.StatusCode, "login must succeed")

	var jwtCookie string
	for _, c := range loginResp.Cookies() {
		if c.Name == "diffgram_jwt" {
			jwtCookie = c.Value
		}
	}
	require.NotEmpty(t, jwtCookie, "diffgram_jwt cookie must be set after login")

	authHeader := fmt.Sprintf("Bearer %s", jwtCookie)

	// 3. Create project
	projBody, _ := json.Marshal(map[string]string{
		"project_name":      "E2E Project",
		"project_string_id": "e2e-project",
	})
	projReq, _ := http.NewRequest(http.MethodPost, base+"/api/project/new", bytes.NewReader(projBody))
	projReq.Header.Set("Content-Type", "application/json")
	projReq.Header.Set("Authorization", authHeader)
	projCreateResp, err := client.Do(projReq)
	require.NoError(t, err)
	defer func() { _ = projCreateResp.Body.Close() }()
	require.Equal(t, http.StatusOK, projCreateResp.StatusCode, "create project must succeed")

	// 4. List projects
	listReq, _ := http.NewRequest(http.MethodPost, base+"/api/v1/project/list", bytes.NewReader([]byte(`{}`)))
	listReq.Header.Set("Content-Type", "application/json")
	listReq.Header.Set("Authorization", authHeader)
	listResp, err := client.Do(listReq)
	require.NoError(t, err)
	defer func() { _ = listResp.Body.Close() }()
	require.Equal(t, http.StatusOK, listResp.StatusCode, "list projects must succeed")

	var listResult map[string]interface{}
	require.NoError(t, json.NewDecoder(listResp.Body).Decode(&listResult))
	projectList, ok := listResult["project_list"].([]interface{})
	require.True(t, ok, "project_list must be an array")
	assert.Len(t, projectList, 1, "must have exactly 1 project")

	// 5. View project
	viewReq, _ := http.NewRequest(http.MethodGet, base+"/api/project/e2e-project/view", nil)
	viewReq.Header.Set("Authorization", authHeader)
	viewResp, err := client.Do(viewReq)
	require.NoError(t, err)
	defer func() { _ = viewResp.Body.Close() }()
	require.Equal(t, http.StatusOK, viewResp.StatusCode, "view project must succeed")

	var viewResult map[string]interface{}
	require.NoError(t, json.NewDecoder(viewResp.Body).Decode(&viewResult))
	project, ok := viewResult["project"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "e2e-project", project["project_string_id"])
}
