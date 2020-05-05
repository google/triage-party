// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package persist provides a persistence layer for the in-memory cache
package persist

import (
	"strings"

	cloudsql "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/mysql"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"k8s.io/klog/v2"
)

// NewCloudSQL returns a new Google Cloud SQL store (MySQL)
func NewCloudSQL(cfg *Config) (*MySQL, error) {
	// DSN that works:
	// $USER:$PASS@tcp($PROJECT/$REGION/$INSTANCE)/$DB"
	dsn, err := mysql.ParseDSN(cfg.Path)
	if err != nil {
		return nil, err
	}

	mcfg := cloudsql.Cfg(dsn.Addr, dsn.User, dsn.Passwd)
	// Strip port
	mcfg.Addr = strings.Split(dsn.Addr, ":")[0]
	mcfg.Addr = strings.Replace(mcfg.Addr, "/", ":", -1)
	mcfg.DBName = dsn.DBName
	mcfg.ParseTime = true
	klog.Infof("mcfg: %#v", mcfg)

	db, err := cloudsql.DialCfg(mcfg)
	if err != nil {
		return nil, err
	}

	dbx := sqlx.NewDb(db, "mysql")
	return &MySQL{db: dbx}, err
}
