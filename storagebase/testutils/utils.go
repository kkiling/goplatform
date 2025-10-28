package testutils

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"

	"github.com/kkiling/goplatform/log"
	"github.com/kkiling/goplatform/storagebase/postgrebase"
	"github.com/kkiling/goplatform/storagebase/sqlitebase"
)

// findUp ищет файл с именем `filename` в текущей и родительских директориях.
// Возвращает полный путь к файлу или пустую строку, если файл не найден.
func findUp(filename string) string {
	dir, err := os.Getwd() // Текущая директория
	if err != nil {
		return ""
	}

	for {
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			return path // Файл найден
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // Достигли корневой директории
		}
		dir = parent // Переходим в родительскую директорию
	}

	return "" // Файл не найден
}

var envPath = ""
var storageSqlite *sqlitebase.Storage = nil
var storagePostgres *postgrebase.Storage = nil

func init() {
	envPath = findUp(".testenv")
	if envPath == "" {
		panic(".testenv not found in current or parent directories")
	}

	// Загружаем переменные из найденного файла
	if err := godotenv.Load(envPath); err != nil {
		panic(err)
	}
}

func getSqliteTestDNS(t *testing.T) string {
	// Загружаем переменные окружения из .testenv файла
	err := godotenv.Overload(envPath)
	require.NoError(t, err, "Failed to load .testenv file")

	return os.Getenv("SQLITE_DSN")
}

func getPostgresqlTestConnString(t *testing.T) string {
	// Загружаем переменные окружения из .testenv файла
	err := godotenv.Overload(envPath)
	require.NoError(t, err, "Failed to load .testenv file")

	return os.Getenv("POSTGRES_CONN_STRING")
}

func SetupSqlTestDB(t *testing.T) *sqlitebase.Storage {
	if storageSqlite != nil {
		return storageSqlite
	}
	// Инициализируем хранилище
	cfg := sqlitebase.Config{
		DSN: getSqliteTestDNS(t), // Берем DSN из переменных окружения
	}
	logger := log.NewLogger(log.DebugLevel)

	s, err := sqlitebase.NewStorage(cfg, logger)
	require.NoError(t, err)

	// Возвращаем хранилище и функцию очистки
	storageSqlite = s
	return s
}

func SetupPostgresqlTestDB(t *testing.T) *postgrebase.Storage {
	if storagePostgres != nil {
		return storagePostgres
	}
	// Инициализируем хранилище
	cfg := postgrebase.Config{
		ConnString: getPostgresqlTestConnString(t), // Берем DSN из переменных окружения
	}

	s, err := postgrebase.NewPgConn(context.Background(), cfg)
	require.NoError(t, err)

	// Возвращаем хранилище и функцию очистки
	storagePostgres = postgrebase.NewStorage(s)

	return storagePostgres
}
