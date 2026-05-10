package mysql

import (
	"testing"

	"github.com/lianjin/campaign-center-api/server/config"
)

func TestInitSkipsWhenMySQLDisabled(t *testing.T) {
	config.Config = &config.Conf{MySQLConfig: &config.MySQL{Enabled: false}}
	db, err := Init()
	if err != nil {
		t.Fatalf("expected no error when mysql is disabled, got %v", err)
	}
	if db != nil {
		t.Fatalf("expected nil db when mysql is disabled")
	}
}
