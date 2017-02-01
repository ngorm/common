package common

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ngorm/ngorm/dialects"
	"github.com/ngorm/ngorm/model"
)

// DefaultForeignKeyNamer contains the default foreign key name generator method
type DefaultForeignKeyNamer struct {
}

var _c dialects.Dialect = &Dialect{}

type Dialect struct {
	DB model.SQLCommon
	DefaultForeignKeyNamer
}

func (Dialect) GetName() string {
	return "common"
}

func (s *Dialect) SetDB(db model.SQLCommon) {
	s.DB = db
}

func (Dialect) BindVar(i int) string {
	return "$$" // ?
}

func (Dialect) Quote(key string) string {
	return fmt.Sprintf(`"%s"`, key)
}

func (Dialect) DataTypeOf(field *model.StructField) (string, error) {
	dataValue, sqlType, size, additionalType :=
		model.ParseFieldStructForDialect(field)

	if sqlType == "" {
		switch dataValue.Kind() {
		case reflect.Bool:
			sqlType = "BOOLEAN"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uintptr:
			if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok {
				sqlType = "INTEGER AUTO_INCREMENT"
			} else {
				sqlType = "INTEGER"
			}
		case reflect.Int64, reflect.Uint64:
			if _, ok := field.TagSettings["AUTO_INCREMENT"]; ok {
				sqlType = "BIGINT AUTO_INCREMENT"
			} else {
				sqlType = "BIGINT"
			}
		case reflect.Float32, reflect.Float64:
			sqlType = "FLOAT"
		case reflect.String:
			if size > 0 && size < 65532 {
				sqlType = fmt.Sprintf("VARCHAR(%d)", size)
			} else {
				sqlType = "VARCHAR(65532)"
			}
		case reflect.Struct:
			if _, ok := dataValue.Interface().(time.Time); ok {
				sqlType = "TIMESTAMP"
			}
		default:
			if _, ok := dataValue.Interface().([]byte); ok {
				if size > 0 && size < 65532 {
					sqlType = fmt.Sprintf("BINARY(%d)", size)
				} else {
					sqlType = "BINARY(65532)"
				}
			}
		}
	}

	if sqlType == "" {
		return "", fmt.Errorf("invalid sql type %s (%s) for Dialect", dataValue.Type().Name(), dataValue.Kind().String())
	}

	if strings.TrimSpace(additionalType) == "" {
		return sqlType, nil
	}
	return fmt.Sprintf("%v %v", sqlType, additionalType), nil
}

func (s Dialect) HasIndex(tableName string, indexName string) bool {
	var count int
	s.DB.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.STATISTICS WHERE table_schema = ? AND table_name = ? AND index_name = ?", s.CurrentDatabase(), tableName, indexName).Scan(&count)
	return count > 0
}

func (s Dialect) RemoveIndex(tableName string, indexName string) error {
	_, err := s.DB.Exec(fmt.Sprintf("DROP INDEX %v", indexName))
	return err
}

func (s Dialect) HasForeignKey(tableName string, foreignKeyName string) bool {
	return false
}

func (s Dialect) HasTable(tableName string) bool {
	var count int
	s.DB.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.TABLES WHERE table_schema = ? AND table_name = ?", s.CurrentDatabase(), tableName).Scan(&count)
	return count > 0
}

func (s Dialect) HasColumn(tableName string, columnName string) bool {
	var count int
	s.DB.QueryRow("SELECT count(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE table_schema = ? AND table_name = ? AND column_name = ?", s.CurrentDatabase(), tableName, columnName).Scan(&count)
	return count > 0
}

func (s Dialect) CurrentDatabase() (name string) {
	s.DB.QueryRow("SELECT DATABASE()").Scan(&name)
	return
}

func (Dialect) LimitAndOffsetSQL(limit, offset interface{}) (sql string) {
	if limit != nil {
		if parsedLimit, err := strconv.ParseInt(fmt.Sprint(limit), 0, 0); err == nil && parsedLimit > 0 {
			sql += fmt.Sprintf(" LIMIT %d", parsedLimit)
		}
	}
	if offset != nil {
		if parsedOffset, err := strconv.ParseInt(fmt.Sprint(offset), 0, 0); err == nil && parsedOffset > 0 {
			sql += fmt.Sprintf(" OFFSET %d", parsedOffset)
		}
	}
	return
}

func (Dialect) SelectFromDummyTable() string {
	return ""
}

func (Dialect) LastInsertIDReturningSuffix(tableName, columnName string) string {
	return ""
}

func (Dialect) PrimaryKey(keys []string) string {
	return ""
}

func (Dialect) QueryFieldName(name string) string {
	return name + "."
}
func (DefaultForeignKeyNamer) BuildForeignKeyName(tableName, field, dest string) string {
	keyName := fmt.Sprintf("%s_%s_%s_foreign", tableName, field, dest)
	keyName = regexp.MustCompile("(_*[^a-zA-Z]+_*|_+)").ReplaceAllString(keyName, "_")
	return keyName
}
