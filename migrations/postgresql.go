package migrations

import (
	"database/sql"
	"fmt"
	"strings"
)

type postgresqlProvider struct{}

type postgresqlBlueprint struct {
	*blueprint
}

func (p *postgresqlProvider) Create(db *sql.DB, tableName string, callback func(PostgreSQLBlueprint)) error {
	bp := &postgresqlBlueprint{newBlueprint(tableName, db)}
	callback(bp)
	
	// Create table
	createSQL := bp.toCreateTableSQL()
	if _, err := db.Exec(createSQL); err != nil {
		return err
	}
	
	// Create indexes
	for _, indexSQL := range bp.toIndexSQL() {
		if _, err := db.Exec(indexSQL); err != nil {
			return err
		}
	}
	
	// Create foreign keys
	for _, foreignSQL := range bp.toForeignKeySQL() {
		if _, err := db.Exec(foreignSQL); err != nil {
			return err
		}
	}
	
	return nil
}

func (p *postgresqlProvider) Table(db *sql.DB, tableName string, callback func(PostgreSQLBlueprint)) error {
	bp := &postgresqlBlueprint{newBlueprint(tableName, db)}
	callback(bp)
	
	sqls := bp.toAlterSQL()
	for _, sql := range sqls {
		if _, err := db.Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (p *postgresqlProvider) Drop(db *sql.DB, tableName string) error {
	sql := fmt.Sprintf("DROP TABLE \"%s\"", tableName)
	_, err := db.Exec(sql)
	return err
}

func (p *postgresqlProvider) DropIfExists(db *sql.DB, tableName string) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
	_, err := db.Exec(sql)
	return err
}

func (p *postgresqlProvider) HasTable(db *sql.DB, tableName string) (bool, error) {
	query := "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = $1)"
	var exists bool
	err := db.QueryRow(query, tableName).Scan(&exists)
	return exists, err
}

func (p *postgresqlProvider) HasColumn(db *sql.DB, tableName, columnName string) (bool, error) {
	query := "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2)"
	var exists bool
	err := db.QueryRow(query, tableName, columnName).Scan(&exists)
	return exists, err
}

func (bp *postgresqlBlueprint) Serial(name string) ColumnBuilder {
	return bp.AddColumn(name, "SERIAL")
}

func (bp *postgresqlBlueprint) BigSerial(name string) ColumnBuilder {
	return bp.AddColumn(name, "BIGSERIAL")
}

func (bp *postgresqlBlueprint) JSONB(name string) ColumnBuilder {
	return bp.AddColumn(name, "JSONB")
}

func (bp *postgresqlBlueprint) Array(name string, baseType string) ColumnBuilder {
	return bp.AddColumn(name, baseType+"[]")
}

func (bp *postgresqlBlueprint) Inet(name string) ColumnBuilder {
	return bp.AddColumn(name, "INET")
}

func (bp *postgresqlBlueprint) CIDR(name string) ColumnBuilder {
	return bp.AddColumn(name, "CIDR")
}

func (bp *postgresqlBlueprint) MacAddr(name string) ColumnBuilder {
	return bp.AddColumn(name, "MACADDR")
}

func (bp *postgresqlBlueprint) TsVector(name string) ColumnBuilder {
	return bp.AddColumn(name, "TSVECTOR")
}

func (bp *postgresqlBlueprint) XML(name string) ColumnBuilder {
	return bp.AddColumn(name, "XML")
}

func (bp *postgresqlBlueprint) Money(name string) ColumnBuilder {
	return bp.AddColumn(name, "MONEY")
}

func (bp *postgresqlBlueprint) HStore(name string) ColumnBuilder {
	return bp.AddColumn(name, "HSTORE")
}

func (bp *postgresqlBlueprint) ID() ColumnBuilder {
	return bp.AddColumn("id", "BIGSERIAL PRIMARY KEY")
}

func (bp *postgresqlBlueprint) UUID(name string) ColumnBuilder {
	return bp.AddColumn(name, "UUID")
}

func (bp *postgresqlBlueprint) Timestamps() {
	bp.AddColumn("created_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
	bp.AddColumn("updated_at", "TIMESTAMP DEFAULT CURRENT_TIMESTAMP")
}

func (bp *postgresqlBlueprint) toCreateTableSQL() string {
	var parts []string
	
	for _, column := range bp.columns {
		columnSQL := fmt.Sprintf("\"%s\" %s", column.Name, column.Type)
		
		if !column.Nullable && !strings.Contains(column.Type, "SERIAL") {
			columnSQL += " NOT NULL"
		}
		
		if column.Default != nil {
			defaultValue := bp.formatDefaultValue(column.Default)
			columnSQL += fmt.Sprintf(" DEFAULT %s", defaultValue)
		}
		
		parts = append(parts, columnSQL)
	}
	
	return fmt.Sprintf("CREATE TABLE \"%s\" (\n  %s\n)", bp.tableName, strings.Join(parts, ",\n  "))
}

func (bp *postgresqlBlueprint) toIndexSQL() []string {
	var sqls []string
	
	for _, index := range bp.indexes {
		indexSQL := bp.formatIndexSQL(index)
		sqls = append(sqls, indexSQL)
	}
	
	return sqls
}

func (bp *postgresqlBlueprint) toForeignKeySQL() []string {
	var sqls []string
	
	for _, foreign := range bp.foreigns {
		foreignSQL := bp.formatForeignKeySQL(foreign)
		sqls = append(sqls, foreignSQL)
	}
	
	return sqls
}

func (bp *postgresqlBlueprint) toAlterSQL() []string {
	var sqls []string
	
	for _, column := range bp.columns {
		columnSQL := fmt.Sprintf("\"%s\" %s", column.Name, column.Type)
		
		if !column.Nullable && !strings.Contains(column.Type, "SERIAL") {
			columnSQL += " NOT NULL"
		}
		
		if column.Default != nil {
			columnSQL += fmt.Sprintf(" DEFAULT %v", column.Default)
		}
		
		sql := fmt.Sprintf("ALTER TABLE \"%s\" ADD COLUMN %s", bp.tableName, columnSQL)
		sqls = append(sqls, sql)
	}
	
	for _, index := range bp.indexes {
		indexSQL := strings.ReplaceAll(index, "`", "\"")
		sql := fmt.Sprintf("ALTER TABLE \"%s\" ADD %s", bp.tableName, indexSQL)
		sqls = append(sqls, sql)
	}
	
	return sqls
}

func (bp *postgresqlBlueprint) formatDefaultValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		// Special handling for PostgreSQL arrays and JSON
		if v == "{}" {
			return "'{}'"
		}
		if v == "[]" {
			return "'[]'"
		}
		if v == "{user}" {
			return "'{user}'"
		}
		if strings.HasPrefix(v, "{") && strings.HasSuffix(v, "}") {
			return "'" + v + "'"
		}
		return "'" + v + "'"
	case int, int64, float64:
		return fmt.Sprintf("%v", v)
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

func (bp *postgresqlBlueprint) formatIndexSQL(index string) string {
	// Replace MySQL syntax with PostgreSQL
	index = strings.ReplaceAll(index, "`", "\"")
	
	if strings.Contains(index, "INDEX") && !strings.Contains(index, "CREATE") {
		// Convert "INDEX name (columns)" to "CREATE INDEX name ON table (columns)"
		parts := strings.Fields(index)
		if len(parts) >= 3 {
			indexName := parts[1]
			columns := strings.Join(parts[2:], " ")
			return fmt.Sprintf("CREATE INDEX %s ON \"%s\" %s", indexName, bp.tableName, columns)
		}
	}
	
	return index
}

func (bp *postgresqlBlueprint) formatForeignKeySQL(foreign string) string {
	// Replace MySQL syntax with PostgreSQL
	foreign = strings.ReplaceAll(foreign, "`", "\"")
	
	if !strings.Contains(foreign, "ALTER TABLE") {
		return fmt.Sprintf("ALTER TABLE \"%s\" ADD %s", bp.tableName, foreign)
	}
	
	return foreign
}