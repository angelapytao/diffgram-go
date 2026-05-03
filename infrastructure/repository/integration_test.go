package repository_test

import (
	"context"
	"database/sql"
	"path/filepath"
	"runtime"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/gorm"

	"github.com/angelapytao/diffgram-go/domain/entity"
	diffdb "github.com/angelapytao/diffgram-go/infrastructure/db"
	infrarepo "github.com/angelapytao/diffgram-go/infrastructure/repository"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test: requires Docker")
	}

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

	sqlDB, err := sql.Open("mysql", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	_, filename, _, _ := runtime.Caller(0)
	migrationsDir := filepath.Join(filepath.Dir(filename), "..", "..", "migrations")
	require.NoError(t, diffdb.RunMigrations(sqlDB, migrationsDir))

	gdb, err := diffdb.NewConnection(dsn)
	require.NoError(t, err)
	return gdb
}

func TestUserRepository_CRUD(t *testing.T) {
	gdb := setupTestDB(t)
	ctx := context.Background()

	repo := infrarepo.NewUserRepository(gdb)

	email := "test@example.com"
	user := &entity.User{Email: email}
	require.NoError(t, repo.Create(ctx, user))
	assert.Greater(t, user.ID, 0, "ID must be assigned after Create")

	found, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, email, found.Email)

	foundByEmail, err := repo.FindByEmail(ctx, email)
	require.NoError(t, err)
	require.NotNil(t, foundByEmail)
	assert.Equal(t, user.ID, foundByEmail.ID)

	name := "Alice"
	found.FirstName = &name
	require.NoError(t, repo.Save(ctx, found))

	updated, err := repo.FindByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.FirstName)
	assert.Equal(t, "Alice", *updated.FirstName)

	missing, err := repo.FindByID(ctx, 999999)
	require.NoError(t, err)
	assert.Nil(t, missing)
}

func TestProjectRepository_CRUD(t *testing.T) {
	gdb := setupTestDB(t)
	ctx := context.Background()

	repo := infrarepo.NewProjectRepository(gdb)

	sid := "test-proj-001"
	proj := &entity.Project{ProjectStringID: &sid}
	require.NoError(t, repo.Create(ctx, proj))
	assert.Greater(t, proj.ID, 0, "ID must be assigned after Create")

	found, err := repo.FindByID(ctx, proj.ID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, sid, *found.ProjectStringID)

	foundBySID, err := repo.FindByStringID(ctx, sid)
	require.NoError(t, err)
	require.NotNil(t, foundBySID)
	assert.Equal(t, proj.ID, foundBySID.ID)

	list, err := repo.ListByOrgID(ctx, 9999)
	require.NoError(t, err)
	assert.Empty(t, list)

	name := "Test Project"
	found.Name = &name
	require.NoError(t, repo.Save(ctx, found))

	updated, err := repo.FindByID(ctx, proj.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.Name)
	assert.Equal(t, "Test Project", *updated.Name)
}

func TestGormProjectRepo_ListByUserPrimaryID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test: requires Docker")
	}
	ctx := context.Background()
	db := setupTestDB(t)

	userRepo := infrarepo.NewUserRepository(db)
	projectRepo := infrarepo.NewProjectRepository(db)

	// Create owner user — let DB assign ID
	hash := "hashval"
	owner := &entity.User{Email: "listowner@example.com", PasswordHash: &hash}
	require.NoError(t, userRepo.Create(ctx, owner))
	ownerID := owner.ID

	// Create another user so the FK is satisfied for the "other" project
	otherUser := &entity.User{Email: "other@example.com", PasswordHash: &hash}
	require.NoError(t, userRepo.Create(ctx, otherUser))
	otherID := otherUser.ID

	sid1 := "proj-one"
	sid2 := "proj-two"
	require.NoError(t, projectRepo.Create(ctx, &entity.Project{ProjectStringID: &sid1, UserPrimaryID: &ownerID}))
	require.NoError(t, projectRepo.Create(ctx, &entity.Project{ProjectStringID: &sid2, UserPrimaryID: &ownerID}))

	// project belonging to someone else — must not appear in owner's list
	sidOther := "other-proj"
	require.NoError(t, projectRepo.Create(ctx, &entity.Project{ProjectStringID: &sidOther, UserPrimaryID: &otherID}))

	list, err := projectRepo.ListByUserPrimaryID(ctx, ownerID)
	require.NoError(t, err)
	assert.Len(t, list, 2)
}
