package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"

	"github.com/blacked/go-zabbix"
	_ "github.com/lib/pq" // I comment extra_float_digits param in this module. Otherwise it crashes on pgbouncer connection.
)

var (
	zbxServer   = os.Args[1] // Zabbix Server address for zabbix sender
	zbxHost     = os.Args[2] // Zabbix var {HOST.HOST} used in zabbix sender
	pgbHost     = os.Args[3] // Pgbouncer host addr {$PGB_HOST}
	pgbPort     = os.Args[4] // Pgrouncer port number {$PGB_PORT}
	pgbUser     = os.Args[5] // Pgbouncer user {$PGB_USER}
	pgbPassword = os.Args[6] // Pgbouncer password {$PGB_PASSWORD}
	pgbDbname   = os.Args[7] // Pgbouncer stats dbname {$PGB_STAT_DB}
	command     = os.Args[8] // Command [lld,getAll]
)

var metrics []*zabbix.Metric

func main() {

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", pgbHost, pgbPort, pgbUser, pgbPassword, pgbDbname)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	switch command {
	case "lld":
		err := lld(db)
		if err != nil {
			log.Fatal(err)
		}
	case "lldUsers":
		err := lldUsers(db)
		if err != nil {
			log.Fatal(err)
		}
	case "getAll":
		queues := []string{"pools", "stats", "databases"}
		// z := zabbix.NewSender(zbxServer, 10051)
		for _, q := range queues {
			packet, err := getData(db, q)
			if err != nil {
				log.Fatal(err)
			}
			// ok we got packet for zabbix sender let's send it
			dataPacket, _ := json.Marshal(packet)
			fmt.Println(string(dataPacket))
			// _, err = z.Send(packet)
			// if err != nil {
			// 	log.Fatal(err)
			// }
		}
		// This value for item indicates good status
		fmt.Println("OK")

	case "getConfig":
		z := zabbix.NewSender(zbxServer, 10051)
		packet, err := getConfig(db)
		if err != nil {
			log.Fatal(err)
		}
		// ok we got packet for zabbix sender let's send it
		// dataPacket, _ := json.Marshal(packet)
		// fmt.Println(string(dataPacket))
		_, err = z.Send(packet)
		if err != nil {
			log.Fatal(err)
		}
		// This value for item indicates good status
		fmt.Println("OK")

	case "getClients":
		z := zabbix.NewSender(zbxServer, 10051)
		packet, err := getClients(db)
		if err != nil {
			log.Fatal(err)
		}
		// ok we got packet for zabbix sender let's send it
		// dataPacket, _ := json.Marshal(packet)
		// fmt.Println(string(dataPacket))
		_, err = z.Send(packet)
		if err != nil {
			log.Fatal(err)
		}
		// This value for item indicates good status
		fmt.Println("OK")

	case "getVer":
		fmt.Println(getVer())
	default:
		fmt.Println("I don't know this command")
	}
}

func lldUsers(db *sql.DB) error {
	rows, err := db.Query("show users")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var rec ShowUsers
	var res UserData
	for rows.Next() {
		err = rows.Scan(&rec.Name, &rec.poolMode)
		if err != nil {
			return err
		}
		res.Data = append(res.Data, rec)
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}
	fmt.Println(string(resJSON))
	return nil

}

// This returns LLD struct for zabbix server
func lldDb(db *sql.DB) error {

	rows, err := db.Query("show pools")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var rec ShowPool
	var res Data
	for rows.Next() {
		err = rows.Scan(&rec.Database, &rec.user, &rec.clActive,
			&rec.clWaiting, &rec.svActive, &rec.svIdle,
			&rec.svUsed, &rec.svTested, &rec.svLogin,
			&rec.maxwait, &rec.maxwaitUs, &rec.poolMode)
		if err != nil {
			return err
		}
		// Skip pgbouncer db
		if rec.Database == "pgbouncer" {
			continue
		}
		res.Data = append(res.Data, rec)
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}
	fmt.Println(string(resJSON))
	return nil
}

func getVer() string {
	out, err := exec.Command("/usr/sbin/pgbouncer", "-V").Output()
	if err != nil {
		panic(err)
	}
	return string(out)
}

func getData(db *sql.DB, queue string) (packet *zabbix.Packet, err error) {
	var key string
	var val *sql.NullString

	rows, err := db.Query("show " + queue)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.NullString)
	}

	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			log.Fatal(err)
		}

		dbName := vals[0].(*sql.NullString) // Dbname is 1st column
		//Skip pgbouncer db
		// if dbName.String == "pgbouncer" {
		// 	continue
		// }

		for idx, colName := range cols[1:] {
			key = fmt.Sprintf("pgbouncer.%v[%v,%v]", queue, dbName.String, colName)
			val = vals[idx+1].(*sql.NullString)
			if val.Valid {
				metrics = append(metrics, zabbix.NewMetric(zbxHost, key, val.String))
			} else {
				metrics = append(metrics, zabbix.NewMetric(zbxHost, key, ""))
			}
		}
	}

	rows.Close()
	packet = zabbix.NewPacket(metrics)
	return packet, nil
}

func getConfig(db *sql.DB) (packet *zabbix.Packet, err error) {
	var key string

	rows, err := db.Query("show config")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.NullString)
	}

	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			log.Fatal(err)
		}

		confKey := vals[0].(*sql.NullString)
		confVal := vals[1].(*sql.NullString)

		// Check for nil values and send
		if confVal.Valid && confKey.Valid {
			key = fmt.Sprintf("pgbouncer.config[%v]", confKey.String)
			metrics = append(metrics, zabbix.NewMetric(zbxHost, key, confVal.String))
		} else {
			continue
		}
	}

	rows.Close()
	packet = zabbix.NewPacket(metrics)
	return packet, nil
}

func getClients(db *sql.DB) (packet *zabbix.Packet, err error) {
	var key string
	count := 0

	rows, err := db.Query("show clients")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.NullString)
	}

	for rows.Next() {
		count++ // Count all rows, that's our client connections
	}
	countS := strconv.Itoa(count)

	key = fmt.Sprintf("pgbouncer.clients[count]")
	metrics = append(metrics, zabbix.NewMetric(zbxHost, key, countS))
	rows.Close()
	packet = zabbix.NewPacket(metrics)
	return packet, nil
}

// This returns LLD struct for zabbix server
func lld(db *sql.DB) error {

	var res Lld
	var rec Databases
	rows, err := db.Query("show pools")
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		log.Fatal(err)
	}

	vals := make([]interface{}, len(cols))
	for i := range cols {
		vals[i] = new(sql.NullString)
	}
	for rows.Next() {
		err = rows.Scan(vals...)
		if err != nil {
			log.Fatal(err)
		}

		db := vals[0].(*sql.NullString)

		// Check for nil values and send
		if db.Valid {
			rec.Database = db.String
			res.Db = append(res.Db, rec)
		} else {
			continue
		}
	}
	resJSON, err := json.Marshal(res)
	if err != nil {
		return err
	}
	fmt.Println(string(resJSON))
	return nil
}
