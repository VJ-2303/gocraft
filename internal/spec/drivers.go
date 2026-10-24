package spec

// GoVersion is the Go toolchain version that generated codebases target.
const GoVersion = "1.27"

// DriverInfo describes the database driver for a dialect.
type DriverInfo struct {
	// Name is the driver name registered with database/sql, e.g. "pgx".
	Name string

	// ModPath is the module path to list in go.mod.
	ModPath string

	// ImportPath is the package whose side effects register the driver.
	ImportPath string

	// ExampleDSN is a sample DSN for the dialect, shown in comments and the
	// generated README.
	ExampleDSN string
}

// drivers maps each SQL dialect to its driver details.
var drivers = map[Database]DriverInfo{
	DBSQLite: {
		Name:       "sqlite",
		ModPath:    "modernc.org/sqlite",
		ImportPath: "modernc.org/sqlite",
		ExampleDSN: "db.sqlite?_pragma=foreign_keys(1)",
	},
	DBPostgres: {
		Name:       "pgx",
		ModPath:    "github.com/jackc/pgx/v5",
		ImportPath: "github.com/jackc/pgx/v5/stdlib",
		ExampleDSN: "user:pass@localhost:5432/example?sslmode=disable&default_query_exec_mode=simple_protocol",
	},
	DBMySQL: {
		Name:       "mysql",
		ModPath:    "github.com/go-sql-driver/mysql",
		ImportPath: "github.com/go-sql-driver/mysql",
		ExampleDSN: "user:pass@tcp(localhost:3306)/example?parseTime=true&multiStatements=true",
	},
}

// migrateTags maps each SQL dialect to its golang-migrate build tag.
var migrateTags = map[Database]string{
	DBSQLite:   "sqlite",
	DBPostgres: "postgres",
	DBMySQL:    "mysql",
}

// BasicAuthHashedPassword is the bcrypt hash (cost 12) of the default
// basic auth password "pa55word".
const BasicAuthHashedPassword = "$2a$12$Oz6h/J3bLOmVNgl.6ivcFORe.oAIGvYOYXavY9VV50T3lMwt4dC.a"
