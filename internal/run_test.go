package internal_test

import (
	"net/http"
	"testing"
	"time"

	"murmapp.caster/internal"
	"github.com/stretchr/testify/require"
)

func TestRun_HealthzOK(t *testing.T) {
	

	// ⏳ Стартуем Run()
	go func() {
		err := internal.Run()
		require.NoError(t, err)
	}()

	// ⏲ Ждём, пока сервер поднимется
	time.Sleep(1 * time.Second)

	// ✅ Проверяем /healthz
	resp, err := http.Get("http://localhost:3999/healthz")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
