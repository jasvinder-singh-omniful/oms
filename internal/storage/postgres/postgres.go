package postgres

import (
	"context"

	"github.com/omniful/go_commons/db/sql/postgres"
	"github.com/omniful/go_commons/log"
	appConfig "github.com/si/internal/config"
)

type Postgres struct {
	Cluster *postgres.DbCluster
}

func NewPostgres(ctx context.Context) *Postgres {
	config := appConfig.GetConfig()
	dbConfig := config.Database
	slavesConfig := config.Slaves

	log.Info("Loading PostgreSQL configuration",
		"host", dbConfig.Host,
		"port", dbConfig.Port,
		"user", dbConfig.Username,
		"db", dbConfig.Database,
	)

	masterConfig := postgres.DBConfig{
		Host:                   dbConfig.Host,
		Port:                   dbConfig.Port,
		Username:               dbConfig.Username,
		Password:               dbConfig.Password,
		Dbname:                 dbConfig.Database,
		MaxOpenConnections:     dbConfig.MaxOpenConnections,
		MaxIdleConnections:     dbConfig.MaxIdleConnections,
		ConnMaxLifetime:        dbConfig.ConnMaxLifetime,
		DebugMode:              dbConfig.DebugMode,
		PrepareStmt:            dbConfig.PrepareStmt,
		SkipDefaultTransaction: dbConfig.SkipDefaultTransaction,
	}

	// 2 slaves (len-2)
	slaves := make([]postgres.DBConfig, 0, len(slavesConfig))
    for i, slaveConfig := range slavesConfig {
        slaveDBConfig := postgres.DBConfig{
            Host:                   slaveConfig.Host,
            Port:                   slaveConfig.Port,
            Username:               slaveConfig.Username,
            Password:               slaveConfig.Password,
            Dbname:                 slaveConfig.Database,
            MaxOpenConnections:     slaveConfig.MaxOpenConnections,
            MaxIdleConnections:     slaveConfig.MaxIdleConnections,
            ConnMaxLifetime:        slaveConfig.ConnMaxLifetime,
            DebugMode:              slaveConfig.DebugMode,
            PrepareStmt:            slaveConfig.PrepareStmt,
            SkipDefaultTransaction: slaveConfig.SkipDefaultTransaction,
        }


        slaves = append(slaves, slaveDBConfig)
        log.InfofWithContext(ctx, "slave db's configured",
            "slave_index", i+1,
            "host", slaveDBConfig.Host,
            "port", slaveDBConfig.Port,
        )
    }

	//fmt.Printf("Type of myInt: %T\n", masterConfig.Port)
	//fmt.Printf("configs of master %v", masterConfig)


	dbCluster := postgres.InitializeDBInstance(masterConfig, &slaves)

	log.Info("PostgreSQL database initialized successfully",
		"host", masterConfig.Host,
		"port", masterConfig.Port,
		"database", masterConfig.Dbname,
	)

	return &Postgres{
		Cluster: dbCluster,
	}
}

// returns database cluster for read/write operations
func (s *Postgres) GetDB() *postgres.DbCluster {
	return s.Cluster
}
