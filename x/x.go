package x

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgraph-io/gru/admin/company"
	"github.com/dgraph-io/gru/dgraph"
	jwt "github.com/dgrijalva/jwt-go"
)

var debug = flag.Bool("debug", false, "Whether to print debug info")

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
	// Filename is like dgraph-0-2016-12-15-20-12.rdf.gz
	s := strings.Split(file.Name(), "-")
	if len(s) != 7 {
		fmt.Println("Can't parse file name format.")
		return
	}
	dateTime := s[2:6]
	// Last string is like 12.rdf.gz
	minutes := s[6][:2]
	dateTime = append(dateTime, minutes)
	t, err := time.Parse(layout, strings.Join(dateTime, "-"))
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
	}
}

func DeleteOldBackups() {
	ticker := time.NewTicker(24 * time.Hour)
	for range ticker.C {
		files, err := ioutil.ReadDir("backup")
		if err != nil {
			fmt.Println("While reading backup directory: ", err)
			continue
		}
		for _, file := range files {
			check(file)
		}
	}
}
