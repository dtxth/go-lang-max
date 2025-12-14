# Refresh Token Improvements

## Проблемы

1. **Race Condition**: Concurrent refresh requests могут создать несколько активных токенов
2. **Отсутствие атомарности**: Между Save() и Revoke() оба токена активны
3. **Нет cleanup**: Expired refresh токены не удаляются автоматически

## Решения

### 1. Атомарная операция refresh в БД

```sql
-- Добавить в refresh_postgres.go
func (r *RefreshPostgres) RefreshAtomic(oldJTI, newJTI string, userID int64, expiresAt time.Time) error {
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Проверяем и отзываем старый токен в одной операции
    var revoked bool
    var oldExpiresAt time.Time
    err = tx.QueryRow(
        `UPDATE refresh_tokens SET revoked = TRUE 
         WHERE jti = $1 AND revoked = FALSE AND expires_at > NOW()
         RETURNING revoked, expires_at`,
        oldJTI,
    ).Scan(&revoked, &oldExpiresAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return domain.ErrTokenExpired
        }
        return err
    }

    // Создаем новый токен
    _, err = tx.Exec(
        `INSERT INTO refresh_tokens (jti, user_id, expires_at) VALUES ($1, $2, $3)`,
        newJTI, userID, expiresAt,
    )
    if err != nil {
        return err
    }

    return tx.Commit()
}
```

### 2. Добавить cleanup для refresh токенов

```go
// В cleanup/refresh_cleanup.go
type RefreshCleanupJob struct {
    repo     domain.RefreshTokenRepository
    interval time.Duration
    logger   *log.Logger
}

func (j *RefreshCleanupJob) Start(ctx context.Context) {
    ticker := time.NewTicker(j.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            j.runCleanup()
        }
    }
}

func (j *RefreshCleanupJob) runCleanup() {
    deleted, err := j.repo.DeleteExpired()
    if err != nil {
        j.logger.Printf("Failed to cleanup expired refresh tokens: %v", err)
        return
    }
    j.logger.Printf("Cleaned up %d expired refresh tokens", deleted)
}
```

### 3. Добавить метод DeleteExpired в repository

```go
// В refresh_postgres.go
func (r *RefreshPostgres) DeleteExpired() (int64, error) {
    result, err := r.db.Exec(
        `DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE`,
    )
    if err != nil {
        return 0, err
    }
    return result.RowsAffected()
}
```

## Тестирование

### Тест на race condition
```go
func TestRefresh_ConcurrentRequests(t *testing.T) {
    // Создать токен
    // Запустить 10 concurrent refresh requests
    // Проверить, что только один успешен
    // Проверить, что старый токен отозван
}
```

### Тест на cleanup
```go
func TestRefreshCleanup_DeletesExpiredTokens(t *testing.T) {
    // Создать expired токены
    // Запустить cleanup
    // Проверить, что токены удалены
}
```