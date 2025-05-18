
# Distributed Database System (Master-Slave with Web GUI)

A distributed database system built with Go and MySQL, implementing a basic master-slave architecture and a web-based interface for interacting with each node.

---

##  Project Structure

```
project-root/
├── master/              # Main master node code (Go)
│   └── main.go
├── slave1/              # Slave 1 node code (Go)
│   └── main.go
├── slave2/              # Slave 2 node code (Go)
│   └── main.go
├── web/                 # Web GUI for interacting with nodes
│   ├── index.html       # GUI for Master
│   ├── index1.html      # GUI for Slave 1
│   └── index2.html      # GUI for Slave 2
└── README.md
```

---

##  Overview

- **Master node (`master/main.go`)** handles all requests (create DB, create table, insert, select, update, delete) and replicates changes to all slave nodes.
- **Slave nodes (`slave1/main.go`, `slave2/main.go`)** receive and execute replication commands from the master.
- **Automatic failover** is supported: if the master goes down, a slave promotes itself to master.
- **Web GUI (`web/index.html`, etc.)** allows user interaction with each node visually.

---

##  Requirements

- Go 1.18+
- MySQL running on all nodes (`127.0.0.1:3306`)
- Git
- (Optional) Node.js (for serving the web GUI)

---

## Configuration

1. MySQL should be running locally on each node.
2. Create a user:

```sql
CREATE USER 'root'@'127.0.0.1' IDENTIFIED BY 'root';
GRANT ALL PRIVILEGES ON *.* TO 'root'@'127.0.0.1' WITH GRANT OPTION;
```

3. Configure these environment variables or hardcode them:
   - `PORT`: HTTP port for each node (e.g. 8001, 8002, 8003)
   - `MYSQL_DSN`: e.g. `root:root@tcp(127.0.0.1:3306)/`

4. In `master/main.go`, update slave addresses:

```go
var slaveAddresses = []string{
    "http://localhost:8002",
    "http://localhost:8003",
}
```

---

##  Run the System

### Master Node

```bash
cd master
go build -o master
./master
```

### Slave 1

```bash
cd slave1
go build -o slave1
PORT=8002 ./slave1
```

### Slave 2

```bash
cd slave2
go build -o slave2
PORT=8003 ./slave2
```

---

##  Run the Web GUI

Option 1: Open manually

- `web/index.html` → Master GUI
- `web/index1.html` → Slave 1 GUI
- `web/index2.html` → Slave 2 GUI

Option 2: Serve with Node.js

```bash
cd web
npx http-server -p 8080
# Open http://localhost:8080
```

---

##  API Examples

### Check Health

```bash
curl http://localhost:8001/ping
# => pong
```

### Create Database

```bash
curl -X POST "http://localhost:8001/createdb?name=testdb"
```

### Create Table

```bash
curl -X POST "http://localhost:8001/createtable?dbname=testdb&table=users&schema=id INT PRIMARY KEY, name VARCHAR(50)"
```

or JSON format:

```bash
curl -X POST "http://localhost:8001/createtable?dbname=testdb&table=users"      -H "Content-Type: application/json"      -d '{"columns":[{"Name":"name","DataType":"VARCHAR(50)"},{"Name":"age","DataType":"INT"}]}'
```

### Insert

```bash
curl -X POST "http://localhost:8001/insert"      -H "Content-Type: application/json"      -d '{"dbname":"testdb","table":"users","records":{"name":"QWxpY2U=","age":30}}'
```

> Note: Strings must be base64-encoded.

### Select

```bash
curl "http://localhost:8001/select?dbname=testdb&table=users"
```

### Update

```bash
curl -X POST "http://localhost:8001/update"      -H "Content-Type: application/json"      -d '{"dbname":"testdb","table":"users","set":"age=31","where":"name='Alice'"}'
```

### Delete

```bash
curl -X POST "http://localhost:8001/delete"      -H "Content-Type: application/json"      -d '{"dbname":"testdb","table":"users","where":"age<25"}'
```

---

##  Auto Master Promotion

- Slaves ping the master regularly.
- If the master is unreachable, the first available slave promotes itself.
- The system can be extended to support Raft for advanced election logic.

---

##  Tips

- Use safe SQL practices.
- Add authentication in production.
- Use HTTPS and secure ports.
- Monitor your system using tools like Grafana or Prometheus.

---

##  Contact

If you have any questions or need help setting this up, feel free to open an issue or reach out.
