package config

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/omniful/go_commons/config"
	"github.com/omniful/go_commons/log"
)

type AppConfig struct {
	Environment string
	Server      ServerConfig
	Database    DatabaseConfig
	Slaves      []DatabaseConfig
}

type ServerConfig struct {
	Host         string
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host                   string
	Port                   string
	Username               string
	Password               string
	Database               string
	MaxOpenConnections     int
	MaxIdleConnections     int
	ConnMaxLifetime        time.Duration
	DebugMode              bool
	PrepareStmt            bool
	SkipDefaultTransaction bool
}

// global instance of AAppConfig
var (
	AppConf *AppConfig
)

func Init(ctx context.Context) error {
	log.InfofWithContext(ctx, "initializing application configuration")

	if err := config.Init(15 * time.Second); err != nil {
		log.ErrorfWithContext(ctx, "failed to initialize go_commons config", err)
		return err
	}

    masterDB := DatabaseConfig{
        Host:                   config.GetString(ctx, "postgres.master.host"),
        Port:                   config.GetString(ctx, "postgres.master.port"),
        Username:               config.GetString(ctx, "postgres.master.user"),
        Password:               config.GetString(ctx, "postgres.master.password"),
        Database:               config.GetString(ctx, "postgres.master.db"),
        MaxOpenConnections:     config.GetInt(ctx, "postgres.master.max_open_connections"),
        MaxIdleConnections:     config.GetInt(ctx, "postgres.master.max_idle_connections"),
        ConnMaxLifetime:        config.GetDuration(ctx, "postgres.master.conn_max_lifetime"),
        DebugMode:              config.GetBool(ctx, "postgres.master.debug_mode"),
        PrepareStmt:            config.GetBool(ctx, "postgres.master.prepare_stmt"),
        SkipDefaultTransaction: config.GetBool(ctx, "postgres.master.skip_default_transaction"),
    }

    slaves := loadSlavesConfig(ctx)

	AppConf = &AppConfig{
        Environment: config.GetString(ctx, "env"),
        Server: ServerConfig{
            Host:         config.GetString(ctx, "http_server.host"),
            Port:         config.GetString(ctx, "http_server.port"),
            ReadTimeout:  config.GetDuration(ctx, "http_server.read_timeout"),
            WriteTimeout: config.GetDuration(ctx, "http_server.write_timeout"),
            IdleTimeout:  config.GetDuration(ctx, "http_server.idle_timeout"),
        },
        Database: masterDB,
        Slaves:   slaves,
    }

	if err := validate(); err != nil {
	    log.WithErrorContext(ctx, fmt.Errorf("configuration validation failed %v", err))
	    return err
	}

	log.InfofWithContext(ctx, "application configuration initialized successfully",
		"environment", AppConf.Environment,
		"server_host", AppConf.Server.Host,
		"server_port", AppConf.Server.Port,
		"db_host", AppConf.Database.Host,
		"db_port", AppConf.Database.Port,
		"db_name", AppConf.Database.Database,
	)

	return nil
}

func loadSlavesConfig(ctx context.Context) []DatabaseConfig {
    slaves := make([]DatabaseConfig, 0)
    
    slaveCount := config.GetInt(ctx, "postgres.slaves.count")
    
    for i := 0; i < slaveCount; i++ {
        slavePrefix := fmt.Sprintf("postgres.slaves.slave_%d", i+1)
        
        slave := DatabaseConfig{
            Host:                   config.GetString(ctx, slavePrefix+".host"),
            Port:                   config.GetString(ctx, slavePrefix+".port"),
            Username:               config.GetString(ctx, slavePrefix+".user"),
            Password:               config.GetString(ctx, slavePrefix+".password"),
            Database:               config.GetString(ctx, slavePrefix+".db"),
            MaxOpenConnections:     config.GetInt(ctx, slavePrefix+".max_open_connections"),
            MaxIdleConnections:     config.GetInt(ctx, slavePrefix+".max_idle_connections"),
            ConnMaxLifetime:        config.GetDuration(ctx, slavePrefix+".conn_max_lifetime"),
            DebugMode:              config.GetBool(ctx, slavePrefix+".debug_mode"),
            PrepareStmt:            config.GetBool(ctx, slavePrefix+".prepare_stmt"),
            SkipDefaultTransaction: config.GetBool(ctx, slavePrefix+".skip_default_transaction"),
        }
        
        if slave.Host != "" {
            slaves = append(slaves, slave)
            log.InfofWithContext(ctx, "slaves dbs configured", 
                "slave_index", i+1, 
                "host", slave.Host, 
                "port", slave.Port)
        }
    }
    
    return slaves
}

func validate() error {
    if AppConf.Database.Host == "" {
        return errors.New("postgres.host - database host is required")
    }
    if AppConf.Database.Port == "" {
        return errors.New("postgres.port - database port is required")
    }
    if AppConf.Database.Username == "" {
        return errors.New("postgres.user - database username is required")
    }
    if AppConf.Database.Database == "" {
        return errors.New("postgres.db - database name is required")
    }

    return nil
}

func GetConfig() *AppConfig {
    return AppConf
}

func GetServerAddress() string {
    if AppConf == nil {
        return ":3000"
    }
    return fmt.Sprintf("%s:%s", AppConf.Server.Host, AppConf.Server.Port)
}
