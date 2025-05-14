package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func allowCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

var db *sql.DB
var isMaster bool = false
var masterAddress string = "http://localhost:8001"

func main() {
	var err error
	db, err = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3307)/")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	http.HandleFunc("/is-master", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{
			"isMaster": isMaster,
		})
	})

	// Read-only endpoints
	http.HandleFunc("/select", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		selectRecords(w, r)
	})

	go checkMasterHealth()
	fmt.Println("Slave2 server running on port 8003...")
	log.Fatal(http.ListenAndServe(":8003", nil))
}

func checkMasterHealth() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		if !isMaster {
			client := &http.Client{Timeout: 5 * time.Second}
			_, err := client.Get(masterAddress + "/ping")
			if err != nil {
				log.Printf("Master is down: %v", err)
				startElection()
			}
		}
	}
}

func startElection() {
	log.Println("Starting election process...")
	time.Sleep(time.Second * 2)

	// Check if there's already a new master
	client := &http.Client{Timeout: 5 * time.Second}
	_, err := client.Get(masterAddress + "/ping")
	if err == nil {
		return // Another node already became master
	}

	// Promote this node to master if it's the highest priority slave
	promotetoMaster()
}

func promotetoMaster() {
	isMaster = true
	masterAddress = "http://localhost:8003"
	log.Println("This node has been promoted to master")

	// Add master endpoints
	http.HandleFunc("/createdb", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		createDB(w, r)
	})

	http.HandleFunc("/dropdb", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		dropDB(w, r)
	})

	http.HandleFunc("/createtable", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		createTable(w, r)
	})

	http.HandleFunc("/insert", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		insertRecord(w, r)
	})

	http.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		updateRecord(w, r)
	})

	http.HandleFunc("/delete", func(w http.ResponseWriter, r *http.Request) {
		allowCORS(w)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		deleteRecord(w, r)
	})
}

func createDB(w http.ResponseWriter, r *http.Request) {
	dbname := r.URL.Query().Get("name")
	if dbname == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("CREATE DATABASE IF NOT EXISTS " + dbname)
	if err != nil {
		http.Error(w, "Failed to create database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Database created successfully"})
}

func dropDB(w http.ResponseWriter, r *http.Request) {
	dbname := r.URL.Query().Get("name")
	if dbname == "" {
		http.Error(w, "Database name is required", http.StatusBadRequest)
		return
	}

	_, err := db.Exec("DROP DATABASE IF EXISTS " + dbname)
	if err != nil {
		http.Error(w, "Failed to drop database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Database dropped successfully"})
}

func createTable(w http.ResponseWriter, r *http.Request) {
	dbname := r.URL.Query().Get("dbname")
	table := r.URL.Query().Get("table")
	schema := r.URL.Query().Get("schema")

	if dbname == "" || table == "" || schema == "" {
		http.Error(w, "All parameters (dbname, table, schema) are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s (%s)", dbname, table, schema)
	_, err := db.Exec(query)
	if err != nil {
		http.Error(w, "Failed to create table: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Table created successfully"})
}

func selectRecords(w http.ResponseWriter, r *http.Request) {
	dbname := r.URL.Query().Get("dbname")
	table := r.URL.Query().Get("table")

	if dbname == "" || table == "" {
		http.Error(w, "Both dbname and table parameters are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("SELECT * FROM %s.%s", dbname, table)
	rows, err := db.Query(query)
	if err != nil {
		http.Error(w, "Failed to query records: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		http.Error(w, "Failed to get columns: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var results []map[string]interface{}
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i := range columns {
			columnPointers[i] = &columns[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			http.Error(w, "Failed to scan row: "+err.Error(), http.StatusInternalServerError)
			return
		}

		row := make(map[string]interface{})
		for i, col := range cols {
			val := columnPointers[i].(*interface{})
			row[col] = *val
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Error during rows iteration: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func insertRecord(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DBName string `json:"dbname"`
		Table  string `json:"table"`
		Values string `json:"values"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.Table == "" || req.Values == "" {
		http.Error(w, "All fields (dbname, table, values) are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("INSERT INTO %s.%s VALUES (%s)", req.DBName, req.Table, req.Values)
	_, err := db.Exec(query)
	if err != nil {
		http.Error(w, "Failed to insert record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Record inserted successfully"})
}

func updateRecord(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DBName string `json:"dbname"`
		Table  string `json:"table"`
		Set    string `json:"set"`
		Where  string `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.Table == "" || req.Set == "" || req.Where == "" {
		http.Error(w, "All fields (dbname, table, set, where) are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s", req.DBName, req.Table, req.Set, req.Where)
	_, err := db.Exec(query)
	if err != nil {
		http.Error(w, "Failed to update record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Record updated successfully"})
}

func deleteRecord(w http.ResponseWriter, r *http.Request) {
	var req struct {
		DBName string `json:"dbname"`
		Table  string `json:"table"`
		Where  string `json:"where"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.DBName == "" || req.Table == "" || req.Where == "" {
		http.Error(w, "All fields (dbname, table, where) are required", http.StatusBadRequest)
		return
	}

	query := fmt.Sprintf("DELETE FROM %s.%s WHERE %s", req.DBName, req.Table, req.Where)
	_, err := db.Exec(query)
	if err != nil {
		http.Error(w, "Failed to delete record: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Record deleted successfully"})
}
