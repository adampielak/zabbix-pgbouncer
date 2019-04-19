package main

import "database/sql"

// ShowPool returns show pool sql query
type ShowPool struct {
	Database  string `json:"{#DATABASE}"`
	user      string
	clActive  string
	clWaiting string
	svActive  string
	svIdle    string
	svUsed    string
	svTested  string
	svLogin   string
	maxwait   string
	maxwaitUs string
	poolMode  string
}

// ShowStats returns show stats sql query
type ShowStats struct {
	Database        string
	totalXactCount  string
	totalQueryCount string
	totalReceived   string
	totalSent       string
	totalXactTime   string
	totalQueryTime  string
	totalWaitTime   string
	avgXactCount    string
	avgQueryCount   string
	avgRecv         string
	avgSent         string
	avgXactTime     string
	avgQueryTime    string
	avgWaitTime     string
}

// ShowDatabases show databases sql query
type ShowDatabases struct {
	name               string
	host               string
	port               string
	database           string
	forceUser          string
	poolSize           string
	reservePool        string
	poolMode           string
	maxConnections     string
	currentConnections string
	paused             string
	disabled           string
}

// ShowClients show clients sql query
type ShowClients struct {
	tType       string
	user        string
	database    string
	state       string
	addr        string
	port        string
	localAddr   string
	localPort   string
	connectTime string
	requestTime string
	wait        string
	waitUs      string
	ptr         string
	link        string
	remotePid   string
	tls         string
}

// ShowUsers returns show users sql query
type ShowUsers struct {
	Name     string `json:"{#USERNAME}"`
	poolMode sql.NullString
}

// ShowStatsTotals returns show stats_totals sql query
type ShowStatsTotals struct {
	Database      string
	xactCount     string
	queryCount    string
	bytesRecieved string
	bytesSent     string
	xactTime      string
	queryTime     string
	waitTime      string
}

// Data represents Database LLD data sended to zabbix-server
type Data struct {
	Data []ShowPool `json:"data"`
}

// UserData represents  Users LLD data sended to zabbix-server
type UserData struct {
	Data []ShowUsers `json:"data"`
}
