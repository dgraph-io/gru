package auth

import "flag"

var Secret = flag.String("secret", "", "Secret used to sign JWT token")
