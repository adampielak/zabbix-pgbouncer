##### Pgbouncer monitoring with zabbix

##### Install

1. Compile `go build`
2. On monitored host:
   * Copy program to /etc/zabbix/scripts/zabbix-pgbouncer
   * Include `userparameter_pgbouncer.conf` in zabbix agent configuration
   * Add monitoring user to pgbouncer configuration. (see [pgbouncer] stats_users)
   * To apply new config without pgbouncer restart, connect to it management interface and type RELOAD (see pgbouncer documentation).
   * Restart zabbix agent
3. On zabbix server:
   * Import template. You need adjust to your version of zabbix-server.
   * Set variables in template macros.
   * Link template to your hosts.


##### Test
   Try to get lld data from your host:

   Example:
`./zabbix-pgbouncer "zabbix.example.com" "pgbouncer.example.com" "pgbouncer.example.com" 5433 zabbix supersecretpassword pgbouncer lld`

  If it complains about `extra_float_digits param`
  You could remove this param in `github.com/lib/pq` and recompile. 
  Or add `ignore_startup_parameters = extra_float_digits` to `[pgbouncer]` section in pgbouncer config (don't forget to reload configuration)

  Try to pass `lld,getAll,getConfig,getVer,getClients` all should return either lld json for zabbix or `OK`.

#### Various versions of pgbouncer
  Monitoring metrics is vary between different versions of pgbouncer. Columns with additional counters were added in newer versions, and some of them could be renamed. Program will handle all versions of pgbouncer and returning items based on their names for this particular version. So you maybe need to adjust template item Prototypes names according to your version.