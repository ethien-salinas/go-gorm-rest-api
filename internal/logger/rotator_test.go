package logger

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDailyRotator_WritesAndRotates(t *testing.T) {
	dir := t.TempDir()
	r, err := NewDailyRotator(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close() })

	_, err = r.Write([]byte("line1\n"))
	require.NoError(t, err)

	today := time.Now().Format("2006-01-02")
	content, err := os.ReadFile(filepath.Join(dir, "app-"+today+".log"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "line1")

	// Simular cambio de día: rotar a un archivo futuro y escribir directamente bajo el lock
	// para evitar que Write() active el chequeo de time.Now() y vuelva al archivo de hoy.
	tomorrow := time.Now().AddDate(0, 0, 1)
	r.mu.Lock()
	require.NoError(t, r.rotate(tomorrow))
	_, werr := r.current.Write([]byte("line2\n"))
	r.mu.Unlock()
	require.NoError(t, werr)

	content2, err := os.ReadFile(filepath.Join(dir, "app-"+tomorrow.Format("2006-01-02")+".log"))
	require.NoError(t, err)
	assert.Contains(t, string(content2), "line2")
}

func TestNewLogger_NoFile(t *testing.T) {
	log, rotator, err := NewLogger(false, "")
	require.NoError(t, err)
	assert.Nil(t, rotator)
	assert.NotNil(t, log)
}

func TestNewLogger_WithFile(t *testing.T) {
	dir := t.TempDir()
	log, rotator, err := NewLogger(true, dir)
	require.NoError(t, err)
	require.NotNil(t, rotator)
	t.Cleanup(func() { _ = rotator.Close() })

	log.Info("test message")

	today := time.Now().Format("2006-01-02")
	content, err := os.ReadFile(filepath.Join(dir, "app-"+today+".log"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "test message")
}
