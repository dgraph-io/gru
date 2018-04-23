package x

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/dgraph-io/gru/admin/company"
	"github.com/dgraph-io/gru/dgraph"
	jwt "github.com/dgrijalva/jwt-go"
)

var (
	debug  = flag.Bool("debug", false, "Whether to print debug info")
	backup = flag.String("backup", "", "Dgraph backup directory path")
)

func Debug(log interface{}) {
	if *debug {
		fmt.Println(log)
	}
}

func StringInSlice(a string, list []string) int {
	for idx, b := range list {
		if b == a {
			return idx
		}
	}
	return -1
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

type Claims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}

func backupDuration() (int, error) {
	c, err := company.Info()
	if err != nil {
		return 0, err
	}
	if c.Backup == 0 {
		return 60, nil
	}
	return c.Backup, nil
}

func Backup() {
	d, err := backupDuration()
	if err != nil {
		fmt.Println(err)
		return
	}

	ticker := time.NewTicker(time.Hour * time.Duration(d))
	for range ticker.C {
		dur, err := backupDuration()
		if err != nil {
			fmt.Println(err)
			continue
		}

		if d != dur {
			ticker = time.NewTicker(time.Hour * time.Duration(d))
		}

		res, err := http.Get(fmt.Sprintf("%v/admin/backup", *dgraph.Server))
		if err != nil || res.StatusCode != http.StatusOK {
			fmt.Println(err)
		}
	}
}

var layout = "2006-01-02-15-04"

func check(file os.FileInfo) {
	fname := file.Name()
	// Filename is like dgraph-0-2016-12-15-20-12.rdf.gz or dgraph-schema-1-2016-12-15-20-12.rdf.gz
	// Length of file name should be atleast 16(for datetime) + 7(.rdf.gz)
	if len(fname) < 23 {
		fmt.Printf("Can't parse file name format: %+v\n", fname)
		return
	}
	// remove .rdf.gz
	fname = fname[:len(fname)-7]
	dateTime := fname[len(fname)-16:]

	t, err := time.Parse(layout, dateTime)
	if err != nil {
		fmt.Println("While parsing backup filename: ", file.Name())
		return
	}

	c, err := company.Info()
	if err != nil {
		fmt.Println(err)
		return
	}

	if time.Now().After(t.Add(time.Duration(c.BackupDays) * 24 * time.Hour)) {
		if err := os.Remove(fmt.Sprintf("backup/%v", file.Name())); err != nil {
			fmt.Println("While removing file with name: ", err)
		}
		fmt.Println("Deleted old backup file: ", fname)
	}
}

func deleteOldBackups() {
	files, err := ioutil.ReadDir(*backup)
	if err != nil {
		fmt.Println("While reading backup directory: ", err)
		return
	}
	for _, file := range files {
		check(file)
	}
}

func DeleteOldBackups() {
	ticker := time.NewTicker(24 * time.Hour)
	deleteOldBackups()
	for range ticker.C {
		deleteOldBackups()
	}
}
