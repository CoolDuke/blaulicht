package config

import (
    "fmt"
    "errors"
    "io/ioutil"
    "encoding/base64"

    "gopkg.in/yaml.v2"
    "github.com/op/go-logging"
    "github.com/sethvargo/go-password/password"
)

type Config struct {
  ListenAddress          string                       `yaml:"listenAddress"`
  SerialPortDevice       string                       `yaml:"serialPortDevice"`
  SerialPortBaudRate     int                          `yaml:"serialPortBaudRate"`
  LdapEnabled            bool                         `yaml:"ldapEnabled"`
  LdapHosts              []string                     `yaml:"ldapHosts"`
  LdapBaseDn             string                       `yaml:"ldapBaseDn"`
  LdapUserSearchFilter   string                       `yaml:"ldapUserSearchFilter"`
  LdapBindDn             string                       `yaml:"ldapBindDn"`
  LdapBindPassword       string                       `yaml:"ldapBindPassword"`
  AuthTokenSigningKey    string                       `yaml:"authTokenSigningKey"`
  InitialAdminPassword   string                       `yaml:"initialAdminPassword"`
}

func GetConfig(log *logging.Logger, filename string) (Config, error) {
  log.Noticef("Reading configuration from: %s", filename)

  bytes, err := ioutil.ReadFile(filename)
  if err != nil {
    return Config{}, err
  }

  //declare default values
  var config Config = Config{
    ListenAddress: ":8080",
  }

  err = yaml.Unmarshal(bytes, &config)
  if err != nil {
    return Config{}, err
  }

  //validate configuration
  err = nil
  if config.SerialPortDevice == "" {
    err = errors.New("serialPortDevice not defined in configuration")
    return config, err
  }
  if config.SerialPortBaudRate < 1 {
    err = errors.New("serialPortBaudRate not defined in configuration")
    return config, err
  }

  if config.LdapEnabled {
    err = nil
    if len(config.LdapHosts) < 1 {
      err = errors.New("ldapHosts not defined in configuration")
      return config, err
    }
    err = nil
    if config.LdapBaseDn == "" {
      err = errors.New("ldapBaseDn not defined in configuration")
      return config, err
    }
    err = nil
    if config.LdapUserSearchFilter == "" {
      err = errors.New("ldapUserSearchFilter not defined in configuration")
      return config, err
    }
    err = nil
    if config.LdapBindDn == "" {
      err = errors.New("ldapBindDn not defined in configuration")
      return config, err
    }
    err = nil
    if config.LdapBindPassword == "" {
      err = errors.New("ldapBindPassword not defined in configuration")
      return config, err
    }
    decodedLdapBindPassword, err := base64.StdEncoding.DecodeString(config.LdapBindPassword)
    if err != nil {
      err = errors.New("Unable to base64-decode ldapBindPassword defined in configuration: " + err.Error())
      return config, err
    }
    config.LdapBindPassword = string(decodedLdapBindPassword)
  } else {
    //If LDAP is disabled and inital password not set, use a random string to prevent illegal logins
    if config.InitialAdminPassword == "" {
      config.InitialAdminPassword, err = password.Generate(32, 10, 0, false, false)
      if err != nil {
        return config, errors.New(fmt.Sprintf("Error while generating an initial password: %s", err.Error()))
      }
    } else {
      decodedInitialAdminPassword, err := base64.StdEncoding.DecodeString(config.InitialAdminPassword)
      if err != nil {
        err = errors.New("Unable to base64-decode InitialAdminPassword defined in configuration: " + err.Error())
        return config, err
      }
      config.InitialAdminPassword = string(decodedInitialAdminPassword)
    }
  }

  if config.AuthTokenSigningKey == "" {
    err = errors.New("authTokenSigningKey not defined in configuration")
    return config, err
  }
  decodedAuthTokenSigningKey, err := base64.StdEncoding.DecodeString(config.AuthTokenSigningKey)
  if err != nil {
    err = errors.New("Unable to base64-decode authTokenSigningKey defined in configuration: " + err.Error())
    return config, err
  }
  config.AuthTokenSigningKey = string(decodedAuthTokenSigningKey)

  return config, nil
}
