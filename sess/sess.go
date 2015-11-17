package sess

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var az = "us-east-1"

var sess *session.Session

// InitSession Set AWS regin for clients
func InitSession() *session.Session {
	if sess == nil {
		sess = session.New(&aws.Config{Region: aws.String(az)})
	}
	return sess
}
