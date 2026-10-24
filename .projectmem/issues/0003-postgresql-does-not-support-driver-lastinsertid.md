# #0003 PostgreSQL does not support driver LastInsertId; UserModel.Insert must use RETURNING id for postgres

- 2026-10-24T10:51:05Z `issue`: PostgreSQL does not support driver LastInsertId; UserModel.Insert must use RETURNING id for postgres [templates/shared/internal/database/users.go.tmpl:33]
- 2026-10-24T10:51:18Z `attempt`: Branch UserModel.Insert with RETURNING id for postgres dialect (worked)
- 2026-10-24T10:51:23Z `fix`: Use RETURNING id for postgres in UserModel.Insert
