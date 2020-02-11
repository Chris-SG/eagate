package user

import (
	"bufio"
	"fmt"
	"github.com/chris-sg/eagate/ea_db"
	"net/http"
	"strings"
	"time"
)

type dbUser struct {
	Name string `db:"account_name"`
	Cookie string `db:"login_cookie"`
	Expiration int64 `db:"cookie_expiration"`
}

func loadCookieFromDb(accountName string) *http.Cookie {
	queryString := fmt.Sprintf(`SELECT * FROM public."eaGateUser" WHERE (account_name) = ('%s')`, accountName)
	rows, err := ea_db.ExecuteQuery(queryString)
	if err != nil {
		return nil
	}
	if rows.Next() {
		var usr dbUser
		err = rows.StructScan(&usr)
		if err != nil {
			return nil
		}
		timeNow := time.Now().UnixNano() / 1000
		if len(usr.Cookie) == 0 || usr.Expiration < timeNow {
			return nil
		}
		rawReq := fmt.Sprintf("GET / HTTP/1.0\r\nCookie: %s\r\n\r\n", usr.Cookie)
		req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(rawReq)))
		if err != nil {
			return nil
		}
		return req.Cookies()[0]
	}
	return nil
}

func writeCookieToDb(accountName string, cookie *http.Cookie) {
	queryString := fmt.Sprintf(`INSERT INTO public."eaGateUser" VALUES ('%s', '%s', %d) ON CONFLICT (account_name) DO UPDATE SET login_cookie = EXCLUDED.login_cookie, cookie_expiration = EXCLUDED.cookie_expiration`, accountName, cookie.String(), cookie.Expires.UnixNano()/1000)
	ea_db.ExecuteInsert(queryString)
}