package quiz

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/dgraph-io/gru/admin/server"
	"github.com/dgraph-io/gru/dgraph"
	minio "github.com/minio/minio-go"
)

var (
	AwsKeyId  = flag.String("keyid", "", "AWS Access Key Id")
	AwsSecret = flag.String("keysecret", "", "AWS Secret Access Key")
	S3bucket  = flag.String("bucket", "", "S3 bucket to upload and fetch resumes from.")
)

func Resume(w http.ResponseWriter, r *http.Request) {
	sr := server.Response{}
	userId, err := validateToken(r)
	if err != nil {
		sr.Write(w, err.Error(), "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseMultipartForm(32 << 20)
	file, _, err := r.FormFile("resume")
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusBadRequest)
		return
	}
	defer file.Close()

	s3Client, err := minio.New("s3.amazonaws.com", *AwsKeyId, *AwsSecret, true)
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	_, err = s3Client.PutObject(*S3bucket, fmt.Sprintf("%v.pdf", userId), file, "application/octet-stream")
	if err != nil {
		sr.Write(w, err.Error(), "", http.StatusInternalServerError)
		return
	}

	m := new(dgraph.Mutation)
	m.Set(`<_uid_:` + userId + `> <resume> "true" .`)
	if _, err = dgraph.SendMutation(m.String()); err != nil {
		sr.Write(w, "", err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
