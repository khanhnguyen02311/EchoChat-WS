package db

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/khanhnguyen02311/EchoChat-WS/components/db/dbmodels"
	conf "github.com/khanhnguyen02311/EchoChat-WS/configurations"
	"github.com/scylladb/gocqlx/v2"
	"time"
)

type ScyllaDB struct {
	Cluster *gocql.ClusterConfig
	Session *gocqlx.Session
	Tables  *dbmodels.ScyllaDBTables
}

func NewScyllaSession() (*ScyllaDB, error) {
	s := ScyllaDB{}
	s.initCluster()
	err := s.initSession()
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (s *ScyllaDB) initCluster() {
	s.Cluster = gocql.NewCluster(conf.SCYLLA_HOST)
	s.Cluster.Port = conf.SCYLLA_PORT
	s.Cluster.Consistency = gocql.LocalOne
	s.Cluster.SerialConsistency = gocql.LocalSerial
	s.Cluster.Timeout = 15 * time.Second
	s.Cluster.Keyspace = conf.SCYLLA_KEYSPACE
	s.Cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())
	//retryPolicy := &gocql.ExponentialBackoffRetryPolicy{
	//	Min:        time.Second,
	//	Max:        10 * time.Second,
	//	NumRetries: 5,
	//}
	s.Tables = dbmodels.InitScyllaDBTables()
}

func (s *ScyllaDB) initSession() error {
	if s.Session != nil && !s.Session.Closed() {
		return nil
	}
	session, err := gocqlx.WrapSession(s.Cluster.CreateSession())
	if err != nil {
		fmt.Println("An error occurred while creating DB session", err.Error())
		return err
	}
	s.Session = &session
	return nil
}

func (s *ScyllaDB) CloseSession() {
	s.Session.Close()
}

func (s *ScyllaDB) Test() {
	// testing gocqlx
	var msgs []dbmodels.Message
	err := s.Session.Query(s.Tables.MessageByGroupTable.SelectAll()).SelectRelease(&msgs)
	if err != nil {
		fmt.Println("An error occurred while getting all messages", err.Error())
		return
	}
	for _, msg := range msgs {
		fmt.Println(msg)
	}
}
