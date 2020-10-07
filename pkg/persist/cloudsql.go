// Copyright 2020 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package persist provides a persistence layer for the in-memory cache
package persist

import (
	"fmt"
	"strings"

	cmysql "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"k8s.io/klog/v2"
)

// NewCloudSQL returns a new Google Cloud SQL store (MySQL)
func NewCloudSQL(cfg Config) (Cacher, error) {
	// This heuristic may be totally wrong. My apologies.
	if strings.Contains(cfg.Path, "(") {
		return newCloudMySQL(cfg)
	}

	return newCloudPostgres(cfg)
}

func newCloudMySQL(cfg Config) (*MySQL, error) {
	// Example DSN: $USER:$PASS@tcp($PROJECT/$REGION/$INSTANCE)/$DB"
	dsn, err := mysql.ParseDSN(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("mysql parse dsn: %w", err)
	}

	mcfg := cmysql.Cfg(dsn.Addr, dsn.User, dsn.Passwd)
	// Strip port
	mcfg.Addr = strings.Split(dsn.Addr, ":")[0]
	mcfg.Addr = strings.Replace(mcfg.Addr, "/", ":", -1)
	mcfg.DBName = dsn.DBName
	mcfg.ParseTime = true
	klog.Infof("mcfg: %#v", mcfg)

	db, err := cmysql.DialCfg(mcfg)
	if err != nil {
		return nil, fmt.Errorf("cloudmysql dialcfg: %w", err)
	}

	dbx := sqlx.NewDb(db, "mysql")
	return &MySQL{db: dbx}, nil
}

func newCloudPostgres(cfg Config) (*Postgres, error) {
	// required for CloudSQL, as the encryption is between the proxy and upstream instead
	if !strings.Contains(cfg.Path, "sslmode=disable") {
		cfg.Path += " sslmode=disable"
	}

	// See https://github.com/GoogleCloudPlatform/cloudsql-proxy/blob/7e668d9ad0ba579372f5142f149a18c38d14a9d0/proxy/dialers/postgres/hook_test.go#L30
	dbx, err := sqlx.Open("cloudsqlpostgres", cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("cloudsqlpostgres open: %w", err)
	}

	klog.Infof("opened cloudsqlpostgres db at %s", cfg.Path)
	return &Postgres{db: dbx}, nil
}
