package api

import (
  "fmt"
  "strings"
  "io/ioutil"
  "net/http"
  "crypto/tls"
  "encoding/json"

  "blaulicht/helpers"

  "gopkg.in/ldap.v3"
  "github.com/dgrijalva/jwt-go"

  //"github.com/kr/pretty"
)

func AuthLogin(w http.ResponseWriter, r *http.Request) {
  var auth AuthData

  b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
    helpers.HttpError(w, http.StatusInternalServerError, "unauthorized", "Unable to read request body: " + err.Error())
    log.Noticef("Login request from %s failed: Unable to read request body: %s", r.RemoteAddr, err.Error())
		return
	}

  err = json.Unmarshal(b, &auth)
  if err != nil {
    helpers.HttpError(w, http.StatusBadRequest, "unauthorized", "Unable to decode JSON input: " + err.Error())
    log.Noticef("Login request from %s failed: Unable to decode JSON input: %s", r.RemoteAddr, err.Error())
		return
  }

	if len(auth.Username) == 0 || len(auth.Password) == 0 {
		helpers.HttpError(w, http.StatusBadRequest, "unauthorized", "Please provide name and password to obtain the token")
    log.Noticef("Login request from %s failed: No username and/or password given", r.RemoteAddr)
		return
	}


  if Conf.LdapEnabled {
    var ldapConn *ldap.Conn
    err = nil
    for _, host := range Conf.LdapHosts {
      ldapConn, err = ldap.DialTLS("tcp", host, &tls.Config{})
      log.Debugf("Trying to connect to LDAP server: %s", host)
      if err == nil {
        break
      }
    }
    if err != nil {
      helpers.HttpError(w, http.StatusInternalServerError, "unauthorized", "Error opening LDAP connection: " + err.Error())
      log.Errorf("Login request of %s from %s failed: Error opening LDAP connection: %s", auth.Username, r.RemoteAddr, err.Error())
      return
    }
    defer ldapConn.Close()

    err = ldapConn.Bind(Conf.LdapBindDn, Conf.LdapBindPassword)
    if err != nil {
      helpers.HttpError(w, http.StatusInternalServerError, "unauthorized", "Unable to bind to LDAP server: " + err.Error())
      log.Errorf("Login request of %s from %s failed: Unable to bind to LDAP server: %s", auth.Username, r.RemoteAddr, err.Error())
      return
    }

    searchRequest := ldap.NewSearchRequest(
      Conf.LdapBaseDn,
      ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
      fmt.Sprintf(Conf.LdapUserSearchFilter, auth.Username),
      []string{"dn"},
      nil,
    )
    sr, err := ldapConn.Search(searchRequest)
    if err != nil {
      helpers.HttpError(w, http.StatusInternalServerError, "unauthorized", "Error while searching for LDAP user: " + err.Error())
      log.Errorf("Login request of %s from %s failed: Error while searching for LDAP user: %s", auth.Username, r.RemoteAddr, err.Error())
      return
    }
    if len(sr.Entries) != 1 {
      helpers.HttpError(w, http.StatusUnauthorized, "unauthorized", "Username and password do not match")
      log.Warningf("Login request of %s from %s failed: User not found", auth.Username, r.RemoteAddr)
      return
    }

    err = ldapConn.Bind(sr.Entries[0].DN, auth.Password)
    if err != nil {
      helpers.HttpError(w, http.StatusUnauthorized, "unauthorized", "Username and password do not match")
      log.Warningf("Login request of %s from %s failed: Wrong password", auth.Username, r.RemoteAddr)
      return
    }
  } else {
    if auth.Username != "admin" {
      helpers.HttpError(w, http.StatusUnauthorized, "unauthorized", "Username and password do not match")
      log.Warningf("Login request of %s from %s failed: LDAP is disabled, default username is admin", auth.Username, r.RemoteAddr)
      return
    }
    if auth.Password != Conf.InitialAdminPassword {
      helpers.HttpError(w, http.StatusUnauthorized, "unauthorized", "Username and password do not match")
      log.Warningf("Login request of %s from %s failed: Password does not match initial password", auth.Username, r.RemoteAddr)
      return
    }
    log.Infof("Token has been generated, you can now unset the initial password in the configuration.", auth.Username, r.RemoteAddr)
  }
  
  token, err := AuthGetToken(auth.Username)
	if err != nil {
		helpers.HttpError(w, http.StatusInternalServerError, "unauthorized", "Error generating JWT token: " + err.Error())
    log.Errorf("Login request of %s from %s failed: Error generating JWT token: %s", auth.Username, r.RemoteAddr, err.Error())
	} else {
		w.Header().Set("Authorization", "Bearer " + token)
    helpers.HttpRespond(w, http.StatusAccepted, "token", token)
    log.Noticef("Login of %s from %s succeeded", auth.Username, r.RemoteAddr)
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if len(tokenString) == 0 {
			helpers.HttpError(w, http.StatusUnauthorized, "unauthorized", "Missing Authorization Header")
      log.Warningf("Request from %s failed: Missing authentication header", r.RemoteAddr)
      return
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)
		claims, err := AuthVerifyToken(tokenString)
		if err != nil {
      helpers.HttpError(w, http.StatusUnauthorized, "unauthorized", "Error verifying JWT token: " + err.Error())
      log.Warningf("Request from %s failed: Error verifying JWT token: %s", r.RemoteAddr, err.Error())
			return
		}
		name := claims.(jwt.MapClaims)["name"].(string)
		role := claims.(jwt.MapClaims)["role"].(string)

		r.Header.Set("name", name)
		r.Header.Set("role", role)

    log.Noticef("Request of %s (role: %s) from %s: %s", name, role, r.RemoteAddr, r.RequestURI)
		next.ServeHTTP(w, r)
	})
}

func AuthGetToken(name string) (string, error) {
	signingKey := []byte(Conf.AuthTokenSigningKey)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": name,
		"role": "admin",
	})
	tokenString, err := token.SignedString(signingKey)
	return tokenString, err
}

func AuthVerifyToken(tokenString string) (jwt.Claims, error) {
	signingKey := []byte(Conf.AuthTokenSigningKey)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	return token.Claims, err
}

type AuthData struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
