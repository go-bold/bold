package migrations

import (
	"database/sql"
	"fmt"
	"strings"
)

type mysqlProvider struct{}

type mysqlBlueprint struct {
	*blueprint
}

func (m *mysqlProvider) Create(db *sql.DB, tableName string, callback func(MySQLBlueprint)) error {
	bp := &mysqlBlueprint{newBlueprint(tableName, db)}
	callback(bp)
	
	sql := bp.toCreateSQL()
	_, err := db.Exec(sql)
	return err
}

func (m *mysqlProvider) Table(db *sql.DB, tableName string, callback func(MySQLBlueprint)) error {
	bp := &mysqlBlueprint{newBlueprint(tableName, db)}
	callback(bp)
	
	sqls := bp.toAlterSQL()
	for _, sql := range sqls {
		if _, err := db.Exec(sql); err != nil {
			return err
		}
	}
	return nil
}

func (m *mysqlProvider) Drop(db *sql.DB, tableName string) error {
	sql := fmt.Sprintf("DROP TABLE `%s`", tableName)
	_, err := db.Exec(sql)
	return err
}

func (m *mysqlProvider) DropIfExists(db *sql.DB, tableName string) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
	_, err := db.Exec(sql)
	return err
}

func (m *mysqlProvider) HasTable(db *sql.DB, tableName string) (bool, error) {
	query := "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
	var count int
	err := db.QueryRow(query, tableName).Scan(&count)
	return count > 0, err
}

func (m *mysqlProvider) HasColumn(db *sql.DB, tableName, columnName string) (bool, error) {
	query := "SELECT COUNT(*) FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?"
	var count int
	err := db.QueryRow(query, tableName, columnName).Scan(&count)
	return count > 0, err
}

func (bp *mysqlBlueprint) Enum(name string, values []string) ColumnBuilder {
	quotedValues := make([]string, len(values))
	for i, v := range values {
		quotedValues[i] = fmt.Sprintf("'%s'", v)
	}
	enumType := fmt.Sprintf("ENUM(%s)", strings.Join(quotedValues, ", "))
	return bp.AddColumn(name, enumType)
}

func (bp *mysqlBlueprint) Set(name string, values []string) ColumnBuilder {
	quotedValues := make([]string, len(values))
	for i, v := range values {
		quotedValues[i] = fmt.Sprintf("'%s'", v)
	}
	setType := fmt.Sprintf("SET(%s)", strings.Join(quotedValues, ", "))
	return bp.AddColumn(name, setType)
}

func (bp *mysqlBlueprint) Point(name string) ColumnBuilder {
	return bp.AddColumn(name, "POINT")
}

func (bp *mysqlBlueprint) Geometry(name string) ColumnBuilder {
	return bp.AddColumn(name, "GEOMETRY")
}

func (bp *mysqlBlueprint) toCreateSQL() string {
	var parts []string
	
	for _, column := range bp.columns {
		columnSQL := fmt.Sprintf("`%s` %s", column.Name, column.Type)
		
		if !column.Nullable {
			columnSQL += " NOT NULL"
		}
		
		if column.Default != nil {
			columnSQL += fmt.Sprintf(" DEFAULT %v", column.Default)
		}
		
		if column.Comment != "" {
			columnSQL += fmt.Sprintf(" COMMENT '%s'", column.Comment)
		}
		
		parts = append(parts, columnSQL)
	}
	
	for _, index := range bp.indexes {
		parts = append(parts, index)
	}
	
	for _, foreign := range bp.foreigns {
		parts = append(parts, foreign)
	}
	
	return fmt.Sprintf("CREATE TABLE `%s` (\n  %s\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		bp.tableName, strings.Join(parts, ",\n  "))
}

func (bp *mysqlBlueprint) toAlterSQL() []string {
	var sqls []string
	
	for _, column := range bp.columns {
		columnSQL := fmt.Sprintf("`%s` %s", column.Name, column.Type)
		
		if !column.Nullable {
			columnSQL += " NOT NULL"
		}
		
		if column.Default != nil {
			columnSQL += fmt.Sprintf(" DEFAULT %v", column.Default)
		}
		
		if column.After != "" {
			columnSQL += fmt.Sprintf(" AFTER `%s`", column.After)
		}
		
		sql := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s", bp.tableName, columnSQL)
		sqls = append(sqls, sql)
	}
	
	for _, index := range bp.indexes {
		sql := fmt.Sprintf("ALTER TABLE `%s` ADD %s", bp.tableName, index)
		sqls = append(sqls, sql)
	}
	
	return sqls
}