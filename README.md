# pgxpool.Pool – Quick Reference

1. SELECT (single row) → `QueryRow`
```
var user domain.User
err := pool.QueryRow(ctx, "SELECT id, email, name FROM users WHERE id = $1", id).Scan(&user.ID, &user.Email, &user.Name)
if errors.Is(err, pgx.ErrNoRows) { ... }
```

2. SELECT (multiple rows) → `Query` + `rows.Next()` + `rows.Scan()`

```
rows, err := pool.Query(ctx, "SELECT id, email, name FROM users")
if err != nil { ... }
defer rows.Close()

var users []domain.User
for rows.Next() {
    var u domain.User
    if err := rows.Scan(&u.ID, &u.Email, &u.Name); err != nil { ... }
    users = append(users, u)
}
```

3. INSERT / UPDATE / DELETE → `Exec` (returns `pgconn.CommandTag`)

```
cmdTag, err := pool.Exec(ctx, "INSERT INTO users (id, email) VALUES ($1, $2)", id, email)
if err != nil { ... }
rowsAffected := cmdTag.RowsAffected()
```

4. INSERT with RETURNING (e.g., generated ID) → `QueryRow`
```
var newID int
err := pool.QueryRow(ctx, "INSERT INTO products (name) VALUES ($1) RETURNING id", name).Scan(&newID)
```

5. Transaction → `pool.Begin(ctx)` then use the returned `pgx.Tx`
```
tx, err := pool.Begin(ctx)
if err != nil { ... }
defer tx.Rollback(ctx) // safe, no-op if committed

// use tx.Exec, tx.QueryRow, etc.
if err := tx.Commit(ctx); err != nil { ... }
```

6. Batch operations → use `pgx.Batch` for efficiency
```
batch := &pgx.Batch{}
batch.Queue("INSERT INTO ...", args1)
batch.Queue("UPDATE ...", args2)
br := pool.SendBatch(ctx, batch)
defer br.Close()

_, err := br.Exec() // first result
_, err = br.Exec()  // second result
```

